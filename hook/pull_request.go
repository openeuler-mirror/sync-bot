package hook

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"sync-bot/git"
	"sync-bot/gitee"
)

func (s *Server) OpenPullRequest(e gitee.PullRequestEvent) {
	owner := e.Repository.Namespace
	repo := e.Repository.Path
	number := e.PullRequest.Number
	title := e.PullRequest.Title
	targetBranch := e.PullRequest.Base.Ref
	logrus.WithFields(logrus.Fields{
		"owner":  owner,
		"repo":   repo,
		"number": number,
		"title":  title,
	}).Infoln("OpenPullRequest")
	if matchTitle(title) {
		logrus.Infoln("Ignore PullRequest created by sync-bot")
		return
	}
	s.replySyncCheck(owner, repo, number, targetBranch)
}

func (s *Server) MergePullRequest(e gitee.PullRequestEvent) {
	owner := e.Repository.Namespace
	repo := e.Repository.Path
	number := e.PullRequest.Number
	logrus.WithFields(logrus.Fields{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}).Infoln("MergePullRequest")

	comments, err := s.GiteeClient.ListPullRequestComments(owner, repo, number)
	if err != nil {
		logrus.Errorln("List PullRequest comments failed", err)
		return
	}
	logrus.WithFields(logrus.Fields{
		"comments": comments,
	}).Infoln("Get all comments")

	// find the last /sync command
	for i, c := range comments {
		comment := comments[len(comments)-1-i]
		user := c.User.Username
		url := c.HTMLURL
		body := comment.Body
		if matchSync(body) {
			logrus.WithFields(logrus.Fields{
				"comment": body,
			}).Infoln("match /sync command")
			_ = s.sync(owner, repo, e.PullRequest, user, url, body)
			return
		}
	}
	logrus.WithFields(logrus.Fields{
		"comments": comments,
	}).Warnln("Not found valid /sync command in pr comments")
}

func (s *Server) pick(owner string, repo string, opt *SyncCmdOption, branchSet map[string]bool, pr gitee.PullRequest,
	title string, body string, firstSha string, lastSha string) ([]syncStatus, error) {
	number := pr.Number
	sourceBranch := pr.Head.Ref
	r, err := s.GitClient.Clone(owner, repo)
	if err != nil {
		logrus.Errorf("Clone %s/%s failed: %v", owner, repo, err)
		return nil, err
	}

	var status []syncStatus
	for _, branch := range opt.branches {
		// branch not in repository
		if ok := branchSet[branch]; !ok {
			status = append(status, syncStatus{
				Name:   branch,
				Status: branchNonExist,
			})
			continue
		}
		tempBranch := fmt.Sprintf("sync-pr%v-%v-to-%v", number, sourceBranch, branch)
		err = r.Checkout("origin/" + branch)
		if err != nil {
			status = append(status, syncStatus{
				Name:   branch,
				Status: err.Error(),
			})
			continue
		}
		err = r.CheckoutNewBranch(tempBranch, true)
		if err != nil {
			status = append(status, syncStatus{
				Name:   branch,
				Status: err.Error(),
			})
			continue
		}
		err = r.FetchPullRequest(number)
		if err != nil {
			status = append(status, syncStatus{
				Name:   branch,
				Status: err.Error(),
			})
			continue
		}
		err = r.CherryPick(firstSha, lastSha, git.Ours)
		if err != nil {
			status = append(status, syncStatus{
				Name:   branch,
				Status: err.Error(),
			})
			continue
		}
		err = r.Push(tempBranch, true)
		if err != nil {
			status = append(status, syncStatus{
				Name:   branch,
				Status: err.Error(),
			})
			continue
		}
		// Wait for the temporary branch to take effect
		sleepyTime := time.Second
		for i := 0; i < 5; i++ {
			_, err = s.GiteeClient.GetBranch(owner, repo, tempBranch)
			if err != nil {
				logrus.WithError(err).Infof("Waiting for branch %s to exist: retrying %d times", tempBranch, i+1)
				time.Sleep(sleepyTime)
				sleepyTime *= 2
				continue
			}
			break
		}
		var url string
		var st string
		// create pull request
		num, err := s.GiteeClient.CreatePullRequest(owner, repo, title, body, tempBranch, branch, true)
		if err != nil {
			logrus.Errorln("Create PullRequest failed:", err)
			st = err.Error()
		} else {
			logrus.Infoln("Create PullRequest:", num)
			st = "Create sync PR"
			url = fmt.Sprintf("https://gitee.com/%v/%v/pulls/%v", owner, repo, num)
		}
		status = append(status, syncStatus{Name: branch, Status: st, PR: url})
	}
	return status, nil
}

func (s *Server) merge(owner string, repo string, opt *SyncCmdOption, branchSet map[string]bool, pr gitee.PullRequest, title string, body string) ([]syncStatus, error) {
	number := pr.Number
	ref := pr.Head.Sha

	var status []syncStatus
	for _, branch := range opt.branches {
		// branch not in repository
		if ok := branchSet[branch]; !ok {
			status = append(status, syncStatus{
				Name:   branch,
				Status: branchNonExist,
			})
			continue
		}
		// create temp branch
		tempBranch := fmt.Sprintf("sync-pr%v-to-%v", number, branch)
		err := s.GiteeClient.CreateBranch(owner, repo, tempBranch, ref)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tempBranch": tempBranch,
			}).Errorln("Create temp branch failed:", err)
			// TODO: check if branch exist
		} else {
			logrus.Infoln("Create temp branch:", branch)
		}
		var url string
		var st string
		// create pull request
		num, err := s.GiteeClient.CreatePullRequest(owner, repo, title, body, tempBranch, branch, true)
		if err != nil {
			logrus.Errorln("Create PullRequest failed:", err)
			st = err.Error()
		} else {
			logrus.Infoln("Create PullRequest:", num)
			st = "Create sync PR"
			url = fmt.Sprintf("https://gitee.com/%v/%v/pulls/%v", owner, repo, num)
		}
		status = append(status, syncStatus{Name: branch, Status: st, PR: url})
	}
	return status, nil
}

func (s *Server) overwrite() bool {
	panic("implement me")
}

func (s *Server) sync(owner string, repo string, pr gitee.PullRequest, user string, url string, command string) error {
	number := pr.Number

	opt, err := parseSyncCommand(command)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"opt": opt,
		}).Errorln("Parse /sync command failed:", err)
		return err
	}

	issues, err := s.GiteeClient.ListPullRequestIssues(owner, repo, number)
	if err != nil {
		logrus.Errorln("List issues in pull request failed:", err)
		return err
	}

	commits, err := s.GiteeClient.ListPullRequestCommits(owner, repo, number)
	if err != nil {
		logrus.Errorln("List commits failed:", err)
		return err
	}
	for i := range commits {
		commits[i].Commit.Message = strings.ReplaceAll(commits[i].Commit.Message, "\n", "<br>")
	}

	// retrieve all branches
	branches, err := s.GiteeClient.GetBranches(owner, repo, false)
	if err != nil {
		logrus.Errorln("List branches failed:", err)
		return err
	}
	branchSet := make(map[string]bool)
	for _, b := range branches {
		branchSet[b.Name] = true
	}

	title := fmt.Sprintf("[sync-bot] PR-%v: %v", number, pr.Title)

	data := struct {
		PR      string
		Issues  []gitee.Issue
		Commits []gitee.PullRequestCommit
	}{
		PR:      pr.HTMLURL,
		Issues:  issues,
		Commits: commits,
	}

	body, err := executeTemplate(syncPRBodyTmpl, data)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tmpl": syncPRBodyTmpl,
			"data": data,
		}).Errorln("Execute template failed:", err)
		return err
	}

	var status []syncStatus
	switch opt.strategy {
	case Pick:
		firstSha := commits[len(commits)-1].Sha
		lastSha := commits[0].Sha
		status, _ = s.pick(owner, repo, opt, branchSet, pr, title, body, firstSha, lastSha)
	case Merge:
		status, _ = s.merge(owner, repo, opt, branchSet, pr, title, body)
	case Overwrite:
		s.overwrite()
	default:
	}

	comment, err := executeTemplate(syncResultTmpl, struct {
		URL        string
		User       string
		Command    string
		SyncStatus []syncStatus
	}{
		URL:        url,
		User:       user,
		Command:    command,
		SyncStatus: status,
	})
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"tmpl": syncResultTmpl,
			"data": data,
		}).Errorln("Execute template failed:", err)
		return err
	}

	err = s.GiteeClient.CreateComment(owner, repo, number, comment)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"owner":   owner,
			"repo":    repo,
			"number":  number,
			"comment": comment,
		}).Errorln("Create comment failed:", err)
	} else {
		logrus.WithFields(logrus.Fields{
			"owner":   owner,
			"repo":    repo,
			"number":  number,
			"comment": comment,
		}).Infoln("Reply sync.")
	}
	return err
}

func (s *Server) ClosePullRequest(e gitee.PullRequestEvent) {
	owner := e.Repository.Namespace
	repo := e.Repository.Path
	number := e.PullRequest.Number

	logrus.WithFields(logrus.Fields{
		"owner":  owner,
		"repo":   repo,
		"number": number,
	}).Infoln("ClosePullRequest")

	// TODO: close issue created by sync-bot, delete temp branch

}

func (s *Server) HandlePullRequestEvent(e gitee.PullRequestEvent) {
	title := e.PullRequest.Title
	switch e.Action {
	case gitee.ActionOpen:
		if matchTitle(title) {
			logrus.WithFields(logrus.Fields{
				"title": title,
			}).Infoln("StateOpen Pull Request which created by sync-bot, ignore it.")
		} else {
			s.OpenPullRequest(e)
		}
	case gitee.ActionMerge:
		if matchTitle(title) {
			logrus.WithFields(logrus.Fields{
				"title": title,
			}).Infoln("Merge Pull Request which created by sync-bot, ignore it.")
		} else {
			s.MergePullRequest(e)
		}
	case gitee.ActionClose:
		if !matchTitle(title) {
			logrus.WithFields(logrus.Fields{
				"title": title,
			}).Infoln("Close Pull Request which not created by sync-bot, ignore it.")
		} else {
			s.ClosePullRequest(e)
		}
	default:
		logrus.Infoln("Ignoring unhandled action:", e.Action)
	}
}
