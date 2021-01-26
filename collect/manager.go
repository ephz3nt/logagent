package collect

import (
	"github.com/sirupsen/logrus"
	"logagent/common"
)

// tasks manager

type manage struct {
	taskList         map[string]*task           // task jobs
	collectEntryList []common.CollectEntry      //
	configChan       chan []common.CollectEntry // wait new config notify chan
}

func (m *manage) watch() {
	// cycle watch
	for {
		// init config chan
		newConf := <-m.configChan
		logrus.Info("configChan: get new config!", newConf)
		// manage tasks
		for _, conf := range newConf {
			// 1. already exist do nothing
			if m.isExist(conf) {
				logrus.Info("task already exists, ignore...")
				continue
			}
			// 2. if new job will create goroutine task
			t := newTask(conf.Topic, conf.Path) // 创建一个收集任务
			err := t.Init()
			if err != nil {
				logrus.Errorf("collect: init collect agent for path: %s failed, err: %v\n", t.path, err)
				continue
			}
			// 收集日志
			m.taskList[t.path] = t // register task
			go t.run()
			logrus.Infof("create collector: %s\n", t.topic)
		}
		// 3. if delete a exist job will stop current goroutine
		// find exists in m.taskList but not in newConf tasks and stop the task
		for k, t := range m.taskList {
			var exist bool
			for _, nc := range newConf {
				if k == nc.Path {
					exist = true
					break
				}
			}
			if !exist {
				// stop this task
				logrus.Infof("config changed, stoping collector: %s\n", t.path)
				delete(m.taskList, k)
				t.cancel()
			}
		}
	}
}

func (m *manage) isExist(cfg common.CollectEntry) bool {
	_, ok := m.taskList[cfg.Path]
	return ok
}

var manager *manage

func Init(etcdConf []common.CollectEntry) (err error) {
	manager = &manage{
		taskList:         make(map[string]*task, 20),
		collectEntryList: etcdConf,
		configChan:       make(chan []common.CollectEntry), // make block channel
	}
	for _, conf := range etcdConf {
		t := newTask(conf.Topic, conf.Path) // 创建一个收集任务
		err = t.Init()
		if err != nil {
			logrus.Errorf("collect: init collect agent for path: %s failed, err: %v\n", t.path, err)
			continue
		}
		// 收集日志
		manager.taskList[t.path] = t // register task
		go t.run()
		logrus.Infof("create collector: %s\n", t.topic)
	}

	go manager.watch() // wait

	return
}

func SendNewConf(newConf []common.CollectEntry) {
	manager.configChan <- newConf
}
