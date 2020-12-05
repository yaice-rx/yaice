package queue

import (
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"sync"
	"time"
)

/**
 工作流程：
	Exchange和Queue绑定key（一个交换机可以和多个信道进行绑定）
*/

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
	exchangeBinds *ExchangeBinds
	// 监听用户手动关闭
	stopC chan struct{}
	//注册接受者
	receiver func([]byte) bool
	//信道
	ch *amqp.Channel
	//Receiver State
	receiverClose chan bool
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
	m.receiverClose = make(chan bool)
	m.closeC = make(chan *amqp.Error, 1)
	m.conn.NotifyClose(m.closeC)
	go m.keepalive()

	return m, nil
}

func (p *MsgQueue) SetExchangeBinds(eb *ExchangeBinds) bool {
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
	p.ch = ch
	return true
}

func (m *MsgQueue) Close() {
	m.mutex.Lock()
	m.receiverClose <- true
	m.stopC <- struct{}{}
	m.state = StateClosed
	m.mutex.Unlock()
}

func (m *MsgQueue) Producer(mandatory bool, body []byte) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.state != StateOpened {
		return fmt.Errorf("MQ: Not initialized, now state is %d", m.State)
	}
	return m.ch.Publish(m.exchangeBinds.Exch.Name, m.exchangeBinds.Binding.RouteKey, mandatory, false, amqp.Publishing{
		ContentType:     "application/json",
		ContentEncoding: "",
		DeliveryMode:    Persistent,
		Priority:        uint8(5),
		Timestamp:       time.Now(),
		Body:            body,
	})
}

func (m *MsgQueue) Consumer(name string, autoAck bool) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.state != StateOpened {
		return fmt.Errorf("MQ: Not initialized, now state is %d", m.State)
	}
	msgs, err := m.ch.Consume(m.exchangeBinds.Binding.Name, name, autoAck,
		// 是否具有排他性
		false,
		// 如果设置为true，表示不能将同一个connection中发送的消息传递给这个connection中的消费者
		false,
		// 队列消费是否阻塞
		false,
		nil)
	if err != nil {
		return fmt.Errorf("MQ: Consumer(%s) consume queue(%s) failed, %v", m.exchangeBinds.Exch.Name, m.exchangeBinds.Binding.Name, err)
	}
	// 启用协和处理消息
	go func() {
		for d := range msgs {
			// 实现我们要实现的逻辑函数
			if m.receiver(d.Body) {
				d.Ack(true)
			}
		}
	}()
	<-m.receiverClose
	return nil
}

func (m MsgQueue) RegisterReceiver(func_ func([]byte) bool) {
	m.receiver = func_
}

func (m *MsgQueue) State() uint8 {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.state
}

/**
 * 监听
 */
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
	}
}

/**
 * 申请绑定交换机
 */
func applyExchangeBinds(ch *amqp.Channel, eb *ExchangeBinds) error {
	if ch == nil {
		return fmt.Errorf("MQ: Nil producer channel")
	}
	if eb.Exch == nil {
		return fmt.Errorf("MQ: Nil exchange found.")
	}
	//创建交换机
	ex := eb.Exch

	// 用于检查交换机是否存在,已经存在不需要重复声明
	err := ch.ExchangeDeclarePassive(ex.Name, ex.RoutingKey, true, false, false, true, nil)
	if err != nil {
		// 注册交换机
		// name:交换机名称,kind:交换机类型,durable:是否持久化,队列存盘,true服务重启后信息不会丢失,影响性能;autoDelete:是否自动删除;
		// noWait:是否非阻塞, true为是,不等待RMQ返回信息;args:参数,传nil即可; internal:是否为内部
		err = ch.ExchangeDeclare(ex.Name, ex.RoutingKey, true, false, false, true, nil)
	}
	//交换机跟channel绑定
	if eb.Binding == nil {
		return fmt.Errorf("MQ: Nil queue found, exchange:%s", ex.Name)
	}
	//创建channel
	if _, err := ch.QueueDeclare(eb.Binding.Name, eb.Binding.Durable, eb.Binding.AutoDelete, eb.Binding.Exclusive, eb.Binding.NoWait, eb.Binding.Args); err != nil {
		return fmt.Errorf("MQ: Declare queue(%s) failed, %v", eb.Binding.Name, err)
	}
	//channel跟exchange绑定
	if err := ch.QueueBind(eb.Binding.Name, eb.Binding.RouteKey, ex.Name, eb.Binding.NoWait, eb.Binding.Args); err != nil {
		return fmt.Errorf("MQ: Bind exchange(%s) <--> queue(%s) failed, %v", ex.Name, eb.Binding.Name, err)
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
