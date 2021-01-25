package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/clientv3"
	"logagent/collect"
	"logagent/common"
	"time"
)

// etcd
var cli *clientv3.Client

func Init(address []string) (err error) {
	cli, err = clientv3.New(clientv3.Config{
		Endpoints:   address,
		DialTimeout: time.Second * 5,
	})
	if err != nil {
		logrus.Errorf("connect etcd failed, err: %v\n", err)
		return
	}
	return

}

func GetConf(key string) (collectEntryList []common.CollectEntry, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	resp, err := cli.Get(ctx, key)
	defer cancel()
	fmt.Println(resp)
	if err != nil {
		logrus.Errorf("cannot get config from etcd by: %s, err: %v\n", key, err)
		return
	}
	if len(resp.Kvs) == 0 {
		logrus.Errorf("empty key: %s\n", key)
		return
	}
	ret := resp.Kvs[0]
	err = json.Unmarshal(ret.Value, &collectEntryList)
	if err != nil {
		logrus.Errorf("json unmarshal failed: %s\n", err)
		return
	}
	return
}

// etcd key monitor
func WatchKey(key string) {
	for {
		wCh := cli.Watch(context.Background(), key)
		for resp := range wCh {
			logrus.Info("new config found!")
			for _, evt := range resp.Events {
				fmt.Printf("type: %s, key: %s, value: %s\n", evt.Type, evt.Kv.Key, evt.Kv.Value)
				var newConf []common.CollectEntry

				err := json.Unmarshal(evt.Kv.Value, &newConf)
				if err != nil {
					logrus.Errorf("json unmarshal failed: %s\n", err)
					continue
				}
				// send new config to tail pkg
				if evt.Type == clientv3.EventTypeDelete {
					logrus.Warnf("etcd delete the key, key:%v\n", evt.Kv.Key)
					collect.SendNewConf(newConf)
				}
			}
		}
	}
}
