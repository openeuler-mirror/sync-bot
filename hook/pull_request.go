package hook

import (
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"sync-bot/git"
	"sync-bot/gitee"
	"sync-bot/util"
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
	s.greeting(owner, repo, number, targetBranch)
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
		if util.MatchSync(body) {
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

func (s *Server) AutoMerge(e gitee.PullRequestEvent) {
	owner := e.Repository.Namespace
	repo := e.Repository.Path
	number := e.PullRequest.Number
	title := e.PullRequest.Title
	sourceBranch := e.PullRequest.Base.Ref
	targetBranch := e.PullRequest.Base.Ref
	mergeable := e.PullRequest.Mergeable
	logger := logrus.WithFields(logrus.Fields{
		"owner":        owner,
		"repo":         repo,
		"number":       number,
		"title":        title,
		"sourceBranch": sourceBranch,
		"targetBranch": targetBranch,
		"mergeable":    mergeable,
	})

	if !mergeable {
		logger.Infoln("The current pull request can not be merge.")
		comment := "The current pull request can not be merge. "
		err := s.GiteeClient.CreateComment(owner, repo, number, comment)
		if err != nil {
			logger.Errorf("Create comment failed: %v", err)
		}
		return
	}
	r, err := s.GitClient.Clone(owner, repo)
	if err != nil {
		logger.Errorf("Clone repository failed: %v", err)
		return
	}
	err = r.FetchPullRequest(number)
	if err != nil {
		logger.Errorf("Fetch pull request failed: %v", err)
		return
	}
	remoteBranch := "origin/" + targetBranch
	err = r.Checkout(remoteBranch)
	if err != nil {
		logger.Errorf("Checkout %v failed: %v", remoteBranch, err)
		return
	}
	err = r.CheckoutNewBranch(targetBranch, true)
	if err != nil {
		logger.Errorf("Checkout %v failed: %v", targetBranch, err)
		return
	}
	pr := fmt.Sprintf("origin/pull/%v", number)
	err = r.Merge(pr, git.MergeFF)
	if err != nil {
		logger.Errorf("Merge failed: %v", err)
		return
	}
	err = r.Push(targetBranch, true)
	if err != nil {
		logger.Errorf("Push %v failed: %v", targetBranch, err)
		return
	}
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

		// pull for big repos by using upstream repos
		if owner == "openeuler" && repo == "kernel" {
			bigRemote := fmt.Sprintf("%s/%s.git", "https://gitee.com", owner+"/"+repo)

			// check remote
			if hasUpstream, _ := r.ListRemote(); !hasUpstream {
				// add remote
				err = r.AddRemote(bigRemote)
				if err != nil {
					status = append(status, syncStatus{
						Name:   branch,
						Status: err.Error(),
					})
					continue
				}
			}

			_ = r.Clean()

			// create branch in fork repo when it exists in origin repo but not exists in fork repo
			// get fork repo's branches
			forkBranches, err := s.GiteeClient.GetBranches("openeuler-sync-bot", repo, false)
			if err != nil {
				status = append(status, syncStatus{
					Name:   branch,
					Status: err.Error(),
				})
				continue
			}

			// create not existed branches
			forkBranchesList := make(map[string]string, len(forkBranches))
			for _, fb := range forkBranches {
				forkBranchesList[fb.Name] = fb.Name
			}

			if _, ok := forkBranchesList[branch]; !ok {
				err = r.FetchUpstream(branch)
				if err != nil {
					status = append(status, syncStatus{
						Name:   branch,
						Status: err.Error(),
					})
					continue
				}

				err = r.CreateBranchAndPushToOrigin(branch, fmt.Sprintf("upstream/%s", branch))
				if err != nil {
					status = append(status, syncStatus{
						Name:   branch,
						Status: err.Error(),
					})
					continue
				}
			}

			// git checkout branch
			err = r.Checkout("origin/" + branch)
			if err != nil {
				status = append(status, syncStatus{
					Name:   branch,
					Status: err.Error(),
				})
				continue
			}

			// git pull
			err = r.FetchUpstream(branch)
			if err != nil {
				status = append(status, syncStatus{
					Name:   branch,
					Status: err.Error(),
				})
				continue
			}

			err = r.MergeUpstream(branch)
			if err != nil {
				status = append(status, syncStatus{
					Name:   branch,
					Status: err.Error(),
				})
				continue
			}

			// git push
			err = r.PushUpstreamToOrigin(branch)
			if err != nil {
				status = append(status, syncStatus{
					Name:   branch,
					Status: err.Error(),
				})
				continue
			}

		} else {
			_ = r.Clean()
			err = r.Checkout("origin/" + branch)
			if err != nil {
				status = append(status, syncStatus{
					Name:   branch,
					Status: err.Error(),
				})
				continue
			}
		}

		tempBranch := fmt.Sprintf("sync-pr%v-%v-to-%v", number, sourceBranch, branch)
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
		err = r.CherryPick(firstSha, lastSha, git.Theirs)
		if err != nil {
			logrus.Errorln("Cherry pick failed:", err.Error())
			status = append(status, syncStatus{
				Name:   branch,
				Status: syncFailed,
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
		var num int
		sleepyTime := time.Second

		if owner == "openeuler" && repo == "kernel" {
			tempBranch = "openeuler-sync-bot:" + tempBranch
		}

		for i := 0; i < 5; i++ {

			num, err = s.GiteeClient.CreatePullRequest(owner, repo, title, body, tempBranch, branch, true)
			if err != nil {
				logrus.WithError(err).Infof("Create pull request: retrying %d times", i+1)
				time.Sleep(sleepyTime)
				sleepyTime *= 2
				continue
			}
			break
		}
		var url string
		var st string
		if err != nil {
			logrus.Errorln("Create PullRequest failed:", err)
			st = err.Error()
		} else {
			logrus.Infoln("Create PullRequest:", num)
			st = createdPR
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

	title := fmt.Sprintf("[sync] PR-%v: %v", number, pr.Title)

	var body string
	var data interface{}
	if owner == "openeuler" && repo == "kernel" {
		data = struct {
			PR   string
			Body string
		}{
			PR:   pr.HTMLURL,
			Body: pr.Body,
		}

		body, err = executeTemplate(syncPRBodyTmplKernel, data)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tmpl": syncPRBodyTmplKernel,
				"data": data,
			}).Errorln("Execute template failed:", err)
			return err
		}
	} else {
		data = struct {
			PR      string
			Issues  []gitee.Issue
			Commits []gitee.PullRequestCommit
		}{
			PR:      pr.HTMLURL,
			Issues:  issues,
			Commits: commits,
		}

		body, err = executeTemplate(syncPRBodyTmpl, data)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"tmpl": syncPRBodyTmpl,
				"data": data,
			}).Errorln("Execute template failed:", err)
			return err
		}
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
		Command:    strings.TrimSpace(command),
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

func (s *Server) ClosePullRequest(owner, repo string, pr gitee.PullRequest) {
	number := pr.Number
	title := pr.Title
	state := pr.State
	sourceBranch := pr.Head.Ref

	logger := logrus.WithFields(logrus.Fields{
		"owner":        owner,
		"repo":         repo,
		"title":        title,
		"number":       number,
		"state":        state,
		"sourceBranch": sourceBranch,
	})
	logger.Infoln("ClosePullRequest")

	r, err := s.GitClient.Clone(owner, repo)
	if err != nil {
		logger.Errorf("Clone repo failed: %v", err)
		return
	}
	if util.MatchSyncBranch(sourceBranch) && r.RemoteBranchExists(sourceBranch) {
		err = r.DeleteRemoteBranch(sourceBranch)
		if err != nil {
			logger.Errorln("Delete source branch failed:", err)
		}
		return
	}
	logger.Warningf("Source branch %v not found.", sourceBranch)
}

func (s *Server) HandlePullRequestEvent(e gitee.PullRequestEvent) {
	title := e.PullRequest.Title
	owner := e.Repository.Namespace
	repo := e.Repository.Path
	number := e.PullRequest.Number
	sourceBranch := e.PullRequest.Head.Ref
	targetBranch := e.PullRequest.Base.Ref

	logger := logrus.WithFields(logrus.Fields{
		"title":        title,
		"owner":        owner,
		"repo":         repo,
		"number":       number,
		"sourceBranch": sourceBranch,
		"targetBranch": targetBranch,
	})

	// TODO: need to be configurable
	// ignoring repo in openeuler
	if owner == "openeuler" && repo != "docs" && repo != "kernel" {
		logger.Infoln("Ignoring repo in openeuler")
		return
	}

	switch e.Action {
	case gitee.ActionOpen:
		if util.MatchTitle(title) {
			go func() {
				// Temporarily circumvent the problem of label openeuler-cla/yes loss
				// Waitting for the openeuler-cal/yes label to be removed before commenting on /check-cla, so delay one seconds here
				time.Sleep(time.Second * 10)
				err := s.GiteeClient.CreateComment(owner, repo, number, "/check-cla")
				if err != nil {
					logger.Warningln("Create comment /check-cla failed:", err)
					return
				}
				logger.Infoln("Create comment /check-cla")
			}()
		} else if util.MatchSyncBranch(targetBranch) {
			s.AutoMerge(e)
		} else {
			s.OpenPullRequest(e)
		}
	case gitee.ActionMerge:
		if util.MatchTitle(title) {
			logger.Infoln("Merge Pull Request which created by sync-bot, ignore it.")
		} else if util.MatchSyncBranch(targetBranch) {
			logger.Infoln("Merge Pull Request to sync branch, ignore it.")
		} else {
			s.MergePullRequest(e)
		}
	case gitee.ActionUpdate:
		if util.MatchSyncBranch(targetBranch) {
			s.AutoMerge(e)
		} else {
			logger.Infoln("Ignoring unhandled action:", e.Action)
		}
	case gitee.ActionClose:
		if util.MatchTitle(title) {
			s.ClosePullRequest(owner, repo, e.PullRequest)
		} else {
			logger.Infoln("Pull request not create by sync-bot, ignoring it.")
		}
	default:
		logger.Infoln("Ignoring unhandled action:", e.Action)
	}
}
