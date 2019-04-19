package disque

import (
	"log"
	"fmt"
	"time"
	"strings"
	"github.com/garyburd/redigo/redis"
	"github.com/EverythingMe/go-disque/disque"
)

/**
 * 必要方法
 */
type QueueClass interface {
	// 连接
	Connect() (error)
	// push
	Push(string) (error)
	// pop
	Pop(string, func(string) bool)
}

type Queue struct {
	QueueConf 	*QueueConf
	Pool 		*disque.Pool   // 连接池
	Conn    	*disque.Client // 冗余
	// Callback 回调函数
}

type QueueConf struct {
	Servers []string	 // ["127.0.0.1:7711", "127.0.0.1:7712"]
	Qname	string		 // 队列name
}

func NewQueue(queueConf *QueueConf) *Queue {
	mq := &Queue{
		QueueConf: queueConf,
		Pool:	nil,
		Conn: 	nil,
	}
	err := mq.initConnection()
	if err != nil {
		return nil
	}
	return mq
}

// 检测重连
func (mq *Queue) initConnection() error {

	if mq.Pool == nil {
		// pool := disque.NewPool(disque.DialFunc(dial), "127.0.0.1:7711", "127.0.0.1:7712")
		pool := disque.NewPool(disque.DialFunc(dial), mq.QueueConf.Servers...)
		mq.Pool = pool
	}

	// 使用新连接
	// 尝试获取连接，尝试次数为len(nodes)
	for {
		client, err := mq.Pool.Get() // 做一次获取连接尝试
		if err != nil {
			log.Println("Connect disque Get client error:", err)
			// panic(err)
			time.Sleep(1 * time.Second)
			continue
		}
		mq.Conn = &client
		break
	}
	// defer client.Close()
	return nil
}

func (mq *Queue) Push(data string) error {

	err := mq.initConnection()
	if err != nil {
		log.Println("Connect disque Get client error:", err)
		mq.initConnection() // 尝试重连
	}

	// defer mq.Conn.Close()

	ja := disque.AddRequest{
		Job: disque.Job{
			Queue: mq.QueueConf.Qname,
			Data:  []byte(data),
		},
		Timeout: time.Millisecond * 100,
	}

	// Add the job to the queue
	id, err := (*mq.Conn).Add(ja) 
	if err != nil {
		// panic(err)
		return err
	}

	// 验证消息id
	if len(id) != 40 && strings.HasPrefix(id, "D-") {
		return fmt.Errorf("Invalid id. got %s", id)
	}
	return nil
}

func (mq *Queue) Pop(){
	
}

func dial(addr string) (redis.Conn, error) {
	return redis.Dial("tcp", addr)
}