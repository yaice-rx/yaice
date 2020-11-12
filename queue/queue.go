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

	consumers []*Consumer
	//生成者
	producers []*Producer
	// 监听用户手动关闭
	stopC chan struct{}
	// MQ状态
	state uint8
}

func New(url string) *MsgQueue {
	return &MsgQueue{
		url:       url,
		producers: make([]*Producer, 0, 1),
		state:     StateClosed,
	}
}

func (m *MsgQueue) Open() (mq *MsgQueue, err error) {
	// 进行Open期间不允许做任何跟连接有关的事情
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.state == StateOpened {
		return m, errors.New("MQ: Had been opened")
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

func (m *MsgQueue) Close() {
	m.mutex.Lock()

	// close producers
	for _, p := range m.producers {
		p.Close()
	}
	m.producers = m.producers[:0]

	// close consumers
	for _, c := range m.consumers {
		c.Close()
	}
	m.consumers = m.consumers[:0]

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
	m.producers = append(m.producers, p)
	return p, nil
}

func (m *MsgQueue) Consumer(name string) (*Consumer, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.state != StateOpened {
		return nil, fmt.Errorf("MQ: Not initialized, now state is %d", m.State)
	}
	c := newConsumer(name, m)
	m.consumers = append(m.consumers, c)
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

func (m *MsgQueue) channel() (*amqp.Channel, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.conn.Channel()
}

func (m MsgQueue) dial() (*amqp.Connection, error) {
	return amqp.Dial(m.url)
}
