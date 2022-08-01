package rabbitmq

import (
	"Points_Mall/config"
	"Points_Mall/dao"
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"log"
	"strconv"
	"strings"
)

type OrderMQ struct {
	RabbitMQ
	channel   *amqp.Channel
	queueName string
	exchange  string
	key       string
}

// NewOrderMQRabbitMQ 获取likeMQ的对应队列。
func NewOrderMQRabbitMQ(queueName string) *OrderMQ {
	likeMQ := &OrderMQ{
		RabbitMQ:  *Rmq,
		queueName: queueName,
	}
	cha, err := likeMQ.conn.Channel()
	likeMQ.channel = cha
	Rmq.failOnErr(err, "获取通道失败")
	return likeMQ
}

func (l *OrderMQ) Publish(message string) {

	_, err := l.channel.QueueDeclare(
		l.queueName,
		//是否持久化
		false,
		//是否为自动删除
		false,
		//是否具有排他性
		false,
		//是否阻塞
		false,
		//额外属性
		nil,
	)
	if err != nil {
		panic(err)
	}

	err1 := l.channel.Publish(
		l.exchange,
		l.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err1 != nil {
		panic(err)
	}

}

func (l *OrderMQ) Consumer() {

	_, err := l.channel.QueueDeclare(l.queueName, false, false, false, false, nil)

	if err != nil {
		panic(err)
	}

	//2、接收消息
	messages, err1 := l.channel.Consume(
		l.queueName,
		//用来区分多个消费者
		"",
		//是否自动应答
		true,
		//是否具有排他性
		false,
		//如果设置为true，表示不能将同一个connection中发送的消息传递给这个connection中的消费者
		false,
		//消息队列是否阻塞
		false,
		nil,
	)
	if err1 != nil {
		panic(err1)
	}

	forever := make(chan bool)
	switch l.queueName {
	case "order_add":
		//增加订单队列
		go l.consumerOrderAdd(messages)
	case "order_del":
		//取消赞消费队列
		go l.consumerOrderDel(messages)

	}

	log.Printf("[*] Waiting for messagees,To exit press CTRL+C")

	<-forever

}

//增加订单
func (l *OrderMQ) consumerOrderAdd(messages <-chan amqp.Delivery) {
	for d := range messages {
		//参数解析
		params := strings.Split(fmt.Sprintf("%s", d.Body), " ")
		goodId, err := strconv.ParseInt(params[0], 10, 64)
		if err != nil {
			log.Println(err.Error())
		}
		orderId, _ := strconv.ParseInt(params[1], 10, 64)
		userId, _ := strconv.ParseInt(params[2], 10, 64)
		for i := 0; i < config.Attempts; i++ {
			flag := false //默认无问题
			//先查有没有这个订单
			var orderData dao.OrderTable
			orderDao := dao.NewOrderDao()
			orderTable, err := orderDao.GetOrderInfoByOrderId(orderId, goodId)
			if err != nil {
				log.Println("get order info err", err.Error())
				flag = true
			} else {
				if orderTable == (dao.OrderTable{}) {
					orderData.Id = orderId
					orderData.GoodId = goodId
					orderData.Cancel = config.ValidOrder
					orderData.UserId = userId
					err := orderDao.InsertOrderTable(orderData)
					if err != nil {
						log.Println("insert order table err", err.Error())
						flag = true
					} else {
						//更新商品表
						goodDao := dao.NewGoodDao()
						goodTable, err := goodDao.SelectGood(goodId)
						if err != nil {
							log.Println("get good table err", err.Error())
							flag = true
						}
						if goodTable.Number <= 0 {
							log.Println("good already empty")
							flag = true

						} else {
							//如果商品还够
							err := goodDao.DecreaseNumber(goodId, 1)
							if err != nil {
								log.Println("decrease good number err", err.Error())
								flag = true
							}
							log.Println("decrease good number suc")
						}

					}
				} else {
					//如果查到了,直接更新
					err := orderDao.UpdateOrderStatus(userId, goodId, config.ValidOrder)
					if err != nil {
						log.Println("update order status err", err.Error())
						flag = true
					}
					log.Println("update order status suc")
				}

			}
			if !flag {
				break //说明成功了
			}

		}

	}

}

//取消订单
func (l *OrderMQ) consumerOrderDel(messages <-chan amqp.Delivery) {
	for d := range messages {
		//参数解析
		params := strings.Split(fmt.Sprintf("%s", d.Body), " ")
		goodId, err := strconv.ParseInt(params[0], 10, 64)
		if err != nil {
			log.Println(err.Error())
		}
		orderId, _ := strconv.ParseInt(params[1], 10, 64)
		userId, _ := strconv.ParseInt(params[2], 10, 64)
		for i := 0; i < config.Attempts; i++ {
			flag := false

			//查询到的结果的cancel==0
			orderDao := dao.NewOrderDao()
			orderTable, err := orderDao.GetOrderInfoByOrderId(orderId, goodId)
			if err != nil {
				log.Println("get order info err", err.Error())
				flag = true
			} else {
				if orderTable == (dao.OrderTable{}) {
					log.Printf(errors.New("can't find data,this action invalid").Error())
					flag = true
				} else {
					//
					//如果查到取消订单确定的状态,增加商品数量
					goodDao := dao.NewGoodDao()
					err := goodDao.IncreaseNumber(goodId, 1)
					if err != nil {
						log.Println("good number incr err", err.Error())
						flag = true
					}
					log.Println("good number incr suc")
					if err := orderDao.UpdateOrderStatus(userId, goodId, config.InvalidOrder); err != nil {
						log.Println("update order status err", err.Error())

					}
					log.Println("update order status suc")
				}

			}
			if !flag {
				break
			}

		}

	}
}

var RmqOrderAdd *OrderMQ
var RmqOrderDel *OrderMQ

// InitOrderRabbitMQ 初始化rabbitMQ连接。
func InitOrderRabbitMQ() {
	RmqOrderAdd = NewOrderMQRabbitMQ("order_add")
	go RmqOrderAdd.Consumer()

	RmqOrderDel = NewOrderMQRabbitMQ("order_del")
	go RmqOrderDel.Consumer()
}
