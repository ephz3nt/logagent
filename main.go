package main

import (
	"fmt"
	"github.com/go-ini/ini"
	"github.com/sirupsen/logrus"
	"logagent/collect"
	"logagent/etcd"
	"logagent/kafka"
)

type Config struct {
	KafkaConfig   `ini:"kafka"`
	CollectConfig `ini:"collect"`
	EtcdConfig    `ini:"etcd"`
}

type KafkaConfig struct {
	Address []string `ini:"address"`
	Topic   string   `ini:"topic"`
	MsgCap  int64    `ini:"msg_cap"`
}

type CollectConfig struct {
	LogfilePath string `ini:"logfile_path"`
}

type EtcdConfig struct {
	Address    []string `ini:"address"`
	CollectKey string   `ini:"collect_key"`
}

func run() {
	select {}
}

func main() {
	// load config
	config := new(Config)
	err := ini.MapTo(config, "config/config.ini")
	if err != nil {
		logrus.Error("load config failed, err: ", err)
		return
	}
	fmt.Printf("%v\n", config.KafkaConfig.Address)
	// connect kafka
	err = kafka.Init(config.KafkaConfig.Address, config.KafkaConfig.MsgCap)
	if err != nil {
		logrus.Error("init kafka failed, err: ", err)
		return
	}
	logrus.Info("init kafka success")
	// init etcd
	fmt.Println(config.EtcdConfig.Address)
	err = etcd.Init(config.EtcdConfig.Address)
	if err != nil {
		logrus.Error("init etcd failed")
		return
	}
	logrus.Info("init etcd success")
	etcdConf, err := etcd.GetConf(config.EtcdConfig.CollectKey)
	if err != nil {
		logrus.Errorf("get config from etcd failed: %v\n", err)
		return
	}
	// create a goroutine watching etcd config.EtcdConfig.CollectKey change
	go etcd.WatchKey(config.EtcdConfig.CollectKey)
	//init collect
	err = collect.Init(etcdConf)
	if err != nil {
		logrus.Error("init collect agent failed")
		return
	}
	logrus.Info("init collect agent success")

	run()
}
