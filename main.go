package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"

	"sync-bot/gitee"
	"sync-bot/hook"
	"sync-bot/secret"

	"github.com/emicklei/go-restful/v3"
	"github.com/sirupsen/logrus"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})
}

type options struct {
	//dryRun        bool   //
	giteeToken    string //
	port          int    //
	webhookSecret string //
}

func (o *options) Validate() error {
	return nil
}

func gatherOptions(fs *flag.FlagSet, args ...string) options {
	var o options
	//fs.BoolVar(&o.dryRun, "dry-run", true, "Dry run for testing. Uses API tokens but does not mutate.")
	fs.StringVar(&o.giteeToken, "gitee-token", "token.conf", "Path to the file containing the Gitee token.")
	fs.IntVar(&o.port, "port", 8765, "Port to listen on.")
	fs.StringVar(&o.webhookSecret, "webhook-secret", "secret.conf", "Path to the file containing the Gitee Webhook secret.")
	_ = fs.Parse(args)
	return o
}

func main() {
	o := gatherOptions(flag.NewFlagSet(os.Args[0], flag.ExitOnError), os.Args[1:]...)
	if err := o.Validate(); err != nil {
		logrus.WithError(err).Fatal("Invalid options")
	}

	err := secret.LoadSecrets([]string{o.giteeToken, o.webhookSecret})
	if err != nil {
		logrus.WithError(err).Fatal("Load secret failed.")
	}

	server := hook.Server{
		GiteeClient: gitee.NewClient(secret.GetGenerator(o.giteeToken)),
		Secret:      secret.GetGenerator(o.webhookSecret),
	}
	restful.Add(server.WebService())
	port := ":" + strconv.Itoa(o.port)
	logrus.WithFields(logrus.Fields{
		"Option": o,
	}).Infoln("Listen...")
	logrus.Fatal(http.ListenAndServe(port, nil))
}
