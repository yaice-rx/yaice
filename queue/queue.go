package queue

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"sync"
	"time"
)

var (
	StateClosed    = uint8(0)
	StateOpened    = uint8(1)
	StateReopening = uint8(2)
)

type MsgQueue struct {
	// RabbitMQ连接的url
	url string
	// 保护内部数据并发读写
	mutex sync.RWMutex
	// RabbitMQ TCP连接
	conn *amqp.Connection
	// RabbitMQ 监听连接错误
	closeC chan *amqp.Error
	// MQ的exchange与其绑定的queues
	exchangeBinds []*ExchangeBinds

	consumer *Consumer
	//生成者
	producer *Producer
	// 监听用户手动关闭
	stopC chan struct{}
	// MQ状态
	state uint8
}

func New(url string) *MsgQueue {
	return &MsgQueue{
		url:   url,
		state: StateClosed,
	}
}

func (m *MsgQueue) Open() (mq *MsgQueue, err error) {
	// 进行Open期间不允许做任何跟连接有关的事情
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.state == StateOpened {
		return m, fmt.Errorf("rabbit 已经启动")
	}

	if m.conn, err = m.dial(); err != nil {
		return m, fmt.Errorf("MQ: Dial err: %v", err)
	}

	m.state = StateOpened
	m.stopC = make(chan struct{})
	m.closeC = make(chan *amqp.Error, 1)
	m.conn.NotifyClose(m.closeC)

	go m.keepalive()

	return m, nil
}

func (p *MsgQueue) SetExchangeBinds(eb []*ExchangeBinds) bool {
	p.mutex.Lock()
	if p.state != StateOpened {
		p.exchangeBinds = eb
	}
	p.mutex.Unlock()

	// 创建并初始化channel
	ch, err := p.channel()
	if err != nil {
		return false
	}
	if err = applyExchangeBinds(ch, p.exchangeBinds); err != nil {
		ch.Close()
		return false
	}
	return true
}

func (m *MsgQueue) Close() {
	m.mutex.Lock()

	// close producers
	m.producer.Close()

	// close consumers
	m.consumer.Close()

	// close mq connection
	select {
	case <-m.stopC:
		// had been closed
	default:
		close(m.stopC)
	}

	m.mutex.Unlock()

	// wait done
	for m.State() != StateClosed {
		time.Sleep(time.Second)
	}
}

func (m *MsgQueue) Producer(name string) (*Producer, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.state != StateOpened {
		return nil, fmt.Errorf("MQ: Not initialized, now state is %d", m.State)
	}
	p := newProducer(name, m)
	return p, nil
}

func (m *MsgQueue) Consumer(name string) (*Consumer, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.state != StateOpened {
		return nil, fmt.Errorf("MQ: Not initialized, now state is %d", m.State)
	}
	c := newConsumer(name, m)
	return c, nil
}

func (m *MsgQueue) State() uint8 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.state
}

func (m *MsgQueue) keepalive() {
	select {
	case <-m.stopC:
		// 正常关闭
		log.Println("[WARN] MQ: Shutdown normally.")
		m.mutex.Lock()
		m.conn.Close()
		m.state = StateClosed
		m.mutex.Unlock()

	case err := <-m.closeC:
		if err == nil {
			log.Println("[ERROR] MQ: Disconnected with MQ, but Error detail is nil")
		} else {
			log.Printf("[ERROR] MQ: Disconnected with MQ, code:%d, reason:%s\n", err.Code, err.Reason)
		}

		// tcp连接中断, 重新连接
		m.mutex.Lock()
		m.state = StateReopening
		m.mutex.Unlock()

		maxRetry := 99999999
		for i := 0; i < maxRetry; i++ {
			time.Sleep(time.Second)
			if _, e := m.Open(); e != nil {
				log.Printf("[ERROR] MQ: Connection recover failed for %d times, %v\n", i+1, e)
				continue
			}
			log.Printf("[INFO] MQ: Connection recover OK. Total try %d times\n", i+1)
			return
		}
		log.Printf("[ERROR] MQ: Try to reconnect to MQ failed over maxRetry(%d), so exit.\n", maxRetry)
	}
}

/**
 * 申请绑定交换机
 */
func applyExchangeBinds(ch *amqp.Channel, exchangeBinds []*ExchangeBinds) error {
	if ch == nil {
		return fmt.Errorf("MQ: Nil producer channel")
	}
	if len(exchangeBinds) <= 0 {
		return fmt.Errorf("MQ: Empty exchangeBinds")
	}
	for _, eb := range exchangeBinds {
		if eb.Exch == nil {
			return fmt.Errorf("MQ: Nil exchange found.")
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

func (m *MsgQueue) channel() (*amqp.Channel, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.conn.Channel()
}

func (m MsgQueue) dial() (*amqp.Connection, error) {
	return amqp.Dial(m.url)
}
