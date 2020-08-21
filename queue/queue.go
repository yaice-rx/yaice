package queue

import "github.com/streadway/amqp"

//定义MQ的结构体
type RabbitMQ struct {
	conn *amqp.Connection
	channel *amqp.Channel
	//队列名称
	QueueName string
	//交换机
	Exchange string
	//key信息
	key string
	//连接信息
	Myurl string
}

type IQueue interface {
	NewQueue(Conf Config,QueueName string,EXchange string,key string)error
	DestoryMQ()
	Producer()
	Consumer()
}

var QueueManager = _InitQueue()

func _InitQueue() IQueue {
	return &RabbitMQ{}
}

func (r *RabbitMQ) NewQueue(Conf Config,QueueName string, EXchange string, key string)error {
	mq := RabbitMQ{
		channel:   nil,
		QueueName: QueueName,
		Exchange:  EXchange,
		key:       key,
		Myurl:     "amqp://"+Conf.UserName+":"+Conf.Password+"@"+Conf.Uri,
	}
	var err error
	mq.conn,err = amqp.Dial(mq.Myurl)
	mq.channel,err = mq.conn.Channel()
	r = &mq
	return err
}

func (r *RabbitMQ) DestoryMQ(){
	r.channel.Close()
	r.conn.Close()
}

func (r *RabbitMQ)Declare(durable, autoDelete, exclusive, noWait bool, args amqp.Table)  {
	ch,err := r.channel.QueueDeclare(r.QueueName,durable,autoDelete,exclusive,noWait,args)
}

func (r *RabbitMQ)Producer(message string)  {
	r.channel.QueueDeclare(message)
}

func (r *RabbitMQ)Consumer()  {

}