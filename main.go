package main

import (
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"github.com/ChimeraCoder/anaconda"
	"github.com/Sirupsen/logrus"
	"github.com/opiuman/middleware"
)

var (
	logger *middleware.Logger
	conf   *config
	stopCh chan struct{}
	api    *anaconda.TwitterApi
)

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logger = middleware.NewLogger("sweeptaken")
	conf = newConf()
	stopCh = make(chan struct{})
	watchSignal()
}

func main() {
	anaconda.SetConsumerKey(conf.Twitter.ConsumerKey)
	anaconda.SetConsumerSecret(conf.Twitter.ConsumerSecret)
	api = anaconda.NewTwitterApi(conf.Twitter.AccessToken, conf.Twitter.TokenSecret)
	streamSweep()
}

func streamSweep() {
	stream := api.PublicStreamFilter(url.Values{
		"track": conf.Tracks,
	})

	defer stream.Stop()
	for {
		select {
		case v := <-stream.C:
			t, ok := v.(anaconda.Tweet)
			if !ok {
				logger.Warningf("received unexpected value of type %T", v)
				continue
			}

			if t.RetweetedStatus != nil {
				continue
			}

			_, err := api.Retweet(t.Id, false)
			if err != nil {
				logger.Errorf("could not retweet %d: %v", t.Id, err)
				continue
			}
			logger.Infof("retweeted %d", t.Id)

		case <-stopCh:
			return
		}

	}
}

func watchSignal() {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		logger.Warnf("recieved %s signal, shutting down... ", <-sigch)
		close(stopCh)
	}()
}
