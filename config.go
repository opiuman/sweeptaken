package main

import (
	"os"

	"io/ioutil"

	"github.com/Sirupsen/logrus"
	"github.com/goibibo/KafkaLogrus"
	"github.com/opiuman/envparser"
	"github.com/vladoatanasov/logrus_amqp"
	"gopkg.in/yaml.v2"
)

type config struct {
	Twitter struct {
		ConsumerKey    string `yaml:"consumerkey"`
		ConsumerSecret string `yaml:"consumersecret"`
		AccessToken    string `yaml:"accesstoken"`
		TokenSecret    string `yaml:"tokensecret"`
	} `yaml:"twitter"`
	Tracks []string `yaml:"tracks"`
	Log    struct {
		Level    string `yaml:"level"`
		File     string `yaml:"file"`
		AMQPHook struct {
			Server   string `yaml:"server"`
			Exchange string `yaml:"exchange"`
			Username string `yaml:"username"`
			Password string `yaml:"password"`
		} `yaml:"amqphook"`
		KafkaHook struct {
			Addresses []string `yaml:"addresses"`
			Topic     string   `yaml:"topic"`
		} `yaml:"kafkahook"`
	} `yaml:"log"`
}

func newConf() *config {
	//conf := &config{}
	conf := parseConfigFile()
	ep := envparser.New("sweeptaken", yaml.Unmarshal)
	err := ep.Parse(conf)
	if err != nil {
		logger.Fatalf("failed to parse config envs -- %s", err)
		return nil
	}
	conf.configLog()
	return conf
}

func parseConfigFile() *config {
	confFile := "sweeptaken_dev.yml"
	if len(os.Args[1:]) > 0 {
		confFile = os.Args[1]
	}
	conf := &config{}

	f, err := os.Open(confFile)
	defer f.Close()
	if err != nil {
		logger.Infof("%s -- let's try to resolve config from envs", err)
		return conf
	}
	d, err := ioutil.ReadAll(f)
	if err != nil {
		logger.Infof("%s -- let's try to resolve config from envs", err)
		return conf
	}
	err = yaml.Unmarshal(d, conf)
	if err != nil {
		logger.Infof("failed to unmarshal config file %s -- %s -- let's try to resolve config from envs", confFile, err)
		return conf
	}
	return conf
}

func (conf *config) configLog() {
	if conf.Log.File != "" {
		f, err := os.OpenFile(conf.Log.File, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err == nil {
			logrus.SetOutput(f)
		} else {
			logger.Warnf("failed to create '%s' log file, use stderr output -- %s", conf.Log.File, err)
		}
	}
	level, err := logrus.ParseLevel(conf.Log.Level)
	if err != nil {
		logger.Warnf("log level '%s' is not supported, use default 'info' -- %s", level, err)
		level = logrus.InfoLevel
	}
	logger.Logger.Level = level

	if conf.Log.AMQPHook.Exchange != "" && conf.Log.AMQPHook.Server != "" {
		hook := logrus_amqp.NewAMQPHook(conf.Log.AMQPHook.Server, conf.Log.AMQPHook.Username,
			conf.Log.AMQPHook.Password, conf.Log.AMQPHook.Exchange, "")
		logger.Logger.Hooks.Add(hook)
	}
	if conf.Log.KafkaHook.Topic != "" {
		khook, err := kafka_logrus.NewHook(conf.Log.KafkaHook.Addresses, conf.Log.KafkaHook.Topic, "", nil)
		if err != nil {
			logger.Error(err)
		}
		logger.Logger.Hooks.Add(khook)
	}
}
