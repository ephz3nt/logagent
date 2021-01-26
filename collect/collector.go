package collect

import (
	"context"
	"github.com/Shopify/sarama"
	"github.com/hpcloud/tail"
	"github.com/sirupsen/logrus"
	"logagent/common"
	"logagent/kafka"
	"strings"
	"time"
)

type task struct {
	path      string
	topic     string
	collector *tail.Tail
	ctx       context.Context
	cancel    context.CancelFunc
}

var configChan chan []common.CollectEntry

func (t *task) Init() (err error) {
	cfg := tail.Config{
		ReOpen:    true,
		Follow:    true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2},
		MustExist: false,
		Poll:      true,
	}
	t.collector, err = tail.TailFile(t.path, cfg)
	if err != nil {
		logrus.Errorf("collect: create collect agent for path: %s failed, err: %v\n",t.path, err)
		return
	}
	// create a goroutine watching new config
	return
}

func (t *task) run() (err error) {
	for {
		select {
		case <-t.ctx.Done():
			logrus.Infof("collector: %v is stopped.\n", t.path)
			return
		case line, ok := <-t.collector.Lines:
			if !ok {
				logrus.Warnf("collect file close reopen, filename: %s, path: %s", t.collector.Filename, t.path)
				time.Sleep(time.Second)
				continue
			}
			// filter empty line
			if len(strings.Trim(line.Text, "\r")) == 0 {
				logrus.Info("empty line, filtered")
				continue
			}
			//
			msg := &sarama.ProducerMessage{}
			msg.Topic = t.topic
			msg.Value = sarama.StringEncoder(line.Text)
			// send msg to kafka msg chan
			kafka.ToMsgChan(msg)
		}

	}
}

func newTask(topic, path string) *task {
	ctx, cancel := context.WithCancel(context.Background())
	t := new(task)
	t.topic = topic
	t.path = path
	t.ctx = ctx
	t.cancel = cancel
	return t

}
