package queue

import (
	"errors"
	"fmt"
	"github.com/streadway/amqp"
)

type Producer struct {
	mq   *MsgQueue
	name string
	// MQ的exchange与其绑定的queues
	exchangeBinds []*ExchangeBinds
	// MQ的会话channel
	ch *amqp.Channel
	// 生产者confirm开关
	enableConfirm bool
}

func newProducer(name string, queue *MsgQueue) *Producer {
	return &Producer{
		name: name,
		mq:   queue,
	}
}

func (p *Producer) Open() error {
	// 条件检测
	if p.mq == nil {
		return errors.New("MQ: Bad producer")
	}
	if len(p.exchangeBinds) <= 0 {
		return errors.New("MQ: No exchangeBinds found. You should SetExchangeBinds before open.")
	}

	// 创建并初始化channel
	ch, err := p.mq.channel()
	if err != nil {
		return fmt.Errorf("MQ: Create channel failed, %v", err)
	}
	if err = applyExchangeBinds(ch, p.exchangeBinds); err != nil {
		ch.Close()
		return err
	}

	p.ch = ch

	// 初始化发送Confirm
	if p.enableConfirm {
		p.confirmC = make(chan amqp.Confirmation, 1) // channel关闭时自动关闭
		p.ch.Confirm(false)
		p.ch.NotifyPublish(p.confirmC)
		if p.confirm == nil {
			p.confirm = newConfirmHelper()
		} else {
			p.confirm.Reset()
		}

		go p.listenConfirm()
	}

	// 初始化Keepalive
	if true {
		p.stopC = make(chan struct{})
		p.closeC = make(chan *amqp.Error, 1) // channel关闭时自动关闭
		p.ch.NotifyClose(p.closeC)

		go p.keepalive()
	}

	return nil
}

func (p *Producer) Close() {

}

/**
 * 申请绑定交换机
 */
func applyExchangeBinds(ch *amqp.Channel, exchangeBinds []*ExchangeBinds) error {
	if ch == nil {
		return errors.New("MQ: Nil producer channel")
	}
	if len(exchangeBinds) <= 0 {
		return errors.New("MQ: Empty exchangeBinds")
	}
	for _, eb := range exchangeBinds {
		if eb.Exch == nil {
			return errors.New("MQ: Nil exchange found.")
		}
		if len(eb.Bindings) <= 0 {
			return fmt.Errorf("MQ: No bindings queue found for exchange(%s)", eb.Exch.Name)
		}
		//创建交换机
		ex := eb.Exch
		if err := ch.ExchangeDeclare(ex.Name, ex.Kind, ex.Durable, ex.AutoDelete, ex.Internal, ex.NoWait, ex.Args); err != nil {
			return fmt.Errorf("MQ: Declare exchange(%s) failed, %v", ex.Name, err)
		}
		//交换机跟channel绑定
		for _, b := range eb.Bindings {
			if b == nil {
				return fmt.Errorf("MQ: Nil binding found, exchange:%s", ex.Name)
			}
			if len(b.Queues) <= 0 {
				return fmt.Errorf("MQ: No queues found for exchange(%s)", ex.Name)
			}
			for _, q := range b.Queues {
				if q == nil {
					return fmt.Errorf("MQ: Nil queue found, exchange:%s", ex.Name)
				}
				//创建channel
				if _, err := ch.QueueDeclare(q.Name, q.Durable, q.AutoDelete, q.Exclusive, q.NoWait, q.Args); err != nil {
					return fmt.Errorf("MQ: Declare queue(%s) failed, %v", q.Name, err)
				}
				//channel跟exchange绑定
				if err := ch.QueueBind(q.Name, b.RouteKey, ex.Name, b.NoWait, b.Args); err != nil {
					return fmt.Errorf("MQ: Bind exchange(%s) <--> queue(%s) failed, %v", ex.Name, q.Name, err)
				}
			}
		}
	}
	return nil
}
