package queue

import (
	"github.com/streadway/amqp"
)

type Producer struct {
	mq   *MsgQueue
	name string
	// MQ的会话channel
	ch *amqp.Channel
	// Producer关闭控制
	stopC chan struct{}
	// 生产者confirm开关
	enableConfirm bool
}

func newProducer(name string, queue *MsgQueue) *Producer {
	return &Producer{
		name: name,
		mq:   queue,
	}
}

func (p *Producer) Close() {

}
