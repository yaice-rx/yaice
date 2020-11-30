package queue

import (
	"github.com/streadway/amqp"
	"time"
)

// 输出方式
var (
	Transient  uint8 = amqp.Transient  //短暂
	Persistent uint8 = amqp.Persistent //持久
)

// ExchangeBinds exchange ==> routeKey ==> queues
type ExchangeBinds struct {
	Exch    *Exchange
	Binding *QueueBinds
}

// Exchange 基于amqp的Exchange配置
type Exchange struct {
	Name         string
	RoutingKey   string     // key名称
	ExchangeType string     // 交换机类型
	Durable      bool       // 是否持久化
	AutoDelete   bool       // 是否自动删除，true表示是。至少有一条绑定才可以触发自动删除，当所有绑定都与交换器解绑后，会自动删除此交换器。
	Internal     bool       // 是否独立 ,true表示是。客户端无法直接发送msg到内部交换器，只有交换器可以发送msg到内部交换器。
	NoWait       bool       // 是否非阻塞
	Args         amqp.Table // default is nil
}

// Queue 基于amqp的Queue配置
type QueueBinds struct {
	Name       string
	RouteKey   string
	Durable    bool // 是否持久化
	AutoDelete bool // 是否自动删除，true为是。至少有一个消费者连接到队列时才可以触发。当所有消费者都断开时，队列会自动删除
	Exclusive  bool // 是否设置排他，true为是。如果设置为排他，则队列仅对首次声明他的连接可见，并在连接断开时自动删除。
	// （注意，这里说的是连接不是信道，相同连接不同信道是可见的）
	NoWait bool // 是否非阻塞
	Args   amqp.Table
}

// 生产者生产的数据格式
type PublishMsg struct {
	ContentType     string // MIME content type
	ContentEncoding string // 编码格式
	DeliveryMode    uint8  // Transient or Persistent
	Priority        uint8  // 0 to 9
	Timestamp       time.Time
	Body            []byte
}
