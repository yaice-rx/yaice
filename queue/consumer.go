package queue

import (
	"fmt"
	"github.com/streadway/amqp"
	"sync"
)

type Consumer struct {
	mq   *MsgQueue
	name string
	// 保护数据并发安全
	mutex sync.RWMutex

	// MQ的会话channel
	ch *amqp.Channel

	// MQ的exchange与其绑定的queues
	exchangeBinds []*ExchangeBinds

	// Oos prefetch
	prefetch int

	// 上层用于接收消费出来的消息的管道
	callback chan<- Delivery

	// 监听会话channel关闭
	closeC chan *amqp.Error
	// Consumer关闭控制
	stopC chan struct{}

	// Consumer状态
	state uint8
}

type Delivery struct {
	amqp.Delivery
}

func newConsumer(name string, queue *MsgQueue) *Consumer {
	return &Consumer{
		name: name,
		mq:   queue,
	}
}

func (c Consumer) Name() string {
	return c.name
}

// CloseChan 该接口仅用于测试使用, 勿手动调用
func (c *Consumer) CloseChan() {
	c.mutex.Lock()
	c.ch.Close()
	c.mutex.Unlock()
}

func (c *Consumer) SetMsgCallback(cb chan<- Delivery) *Consumer {
	c.mutex.Lock()
	c.callback = cb
	c.mutex.Unlock()
	return c
}

// SetQos 设置channel粒度的Qos, prefetch取值范围[0,∞), 默认为0
// 如果想要RoundRobin地进行消费，设置prefetch为1即可
// 注意:在调用Open前设置
func (c *Consumer) SetQos(prefetch int) *Consumer {
	c.mutex.Lock()
	c.prefetch = prefetch
	c.mutex.Unlock()
	return c
}

func (c *Consumer) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	select {
	case <-c.stopC:
		// had been closed
	default:
		close(c.stopC)
	}
}

// notifyErr 向上层抛出错误, 如果error为空表示执行完成.由上层负责关闭channel
func (c *Consumer) consume(opt *ConsumeOption, notifyErr chan<- error) {
	for idx, eb := range c.exchangeBinds {
		if eb == nil {
			notifyErr <- fmt.Errorf("MQ: ExchangeBinds[%d] is nil, consumer(%s)", idx, c.name)
			continue
		}
		for i, b := range eb.Bindings {
			if b == nil {
				notifyErr <- fmt.Errorf("MQ: Binding[%d] is nil, ExchangeBinds[%d], consumer(%s)", i, idx, c.name)
				continue
			}
			for qi, q := range b.Queues {
				if q == nil {
					notifyErr <- fmt.Errorf("MQ: Queue[%d] is nil, ExchangeBinds[%d], Biding[%d], consumer(%s)", qi, idx, i, c.name)
					continue
				}
				delivery, err := c.ch.Consume(q.Name, "", opt.AutoAck, opt.Exclusive, opt.NoLocal, opt.NoWait, opt.Args)
				if err != nil {
					notifyErr <- fmt.Errorf("MQ: Consumer(%s) consume queue(%s) failed, %v", c.name, q.Name, err)
					continue
				}
				go c.deliver(delivery)
			}
		}
	}
	notifyErr <- nil
}

func (c *Consumer) deliver(delivery <-chan amqp.Delivery) {
	for d := range delivery {
		if c.callback != nil {
			c.callback <- Delivery{d}
		}
	}
}

func (c *Consumer) State() uint8 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.state
}
