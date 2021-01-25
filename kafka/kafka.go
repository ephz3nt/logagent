package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

var client sarama.SyncProducer
var msgChan chan *sarama.ProducerMessage

func Init(address []string, messageCap int64) (err error) {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	cfg.Producer.Return.Successes = true

	client, err = sarama.NewSyncProducer(address, cfg)
	if err != nil {
		logrus.Error("kafka: connect failed, err: ", err)
		return err
	}
	// init MsgChan
	msgChan = make(chan *sarama.ProducerMessage, messageCap)
	// start a sendMsg goroutine
	go sendMsg()
	return
}

// load MsgChan send to kafka server
func sendMsg() {
	for {
		msg := <-msgChan
		pid, offset, err := client.SendMessage(msg)
		if err != nil {
			logrus.Errorf("send msg failed, err: %v", err)
			return
		}
		logrus.Infof("%s: sending success, pid: %d, offset: %d", msg.Topic, pid, offset)

	}
}

func ToMsgChan(msg *sarama.ProducerMessage) {
	msgChan <- msg
}
