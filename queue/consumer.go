package queue

type Consumer struct {
	mq   *MsgQueue
	name string
}

func newConsumer(name string, queue *MsgQueue) *Consumer {
	return &Consumer{
		name: name,
		mq:   queue,
	}
}

func (p *Consumer) Close() {

}
