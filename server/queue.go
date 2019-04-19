package server
/**
	作为消息中间存储层

	单机环境：使用本地 list结构即可
  	高可用下：因为各个节点需要数据同步，所以使用pub/sub结构
**/
import (
	"log"
	"websocket/lib"
	"websocket/queue/redis_queue"
	"websocket/queue/list_queue"
)

//队列类实现方法
type QueueClass interface {
	// 连接
	NewQueue()
	// publish
	Pub(string) (error)
	// Subscribe
	Sub() (error)
}

func (ws *Ws) Callback(data string) error {
	d := lib.JsonMsgToSendMsg(data)
	if d == nil {
		log.Println("Queue.Callback.JsonMsgToSendMsg err")
	}
	ws.cl.SendMsg(d)
	return nil
}

// 初始化queue
// 无法做反射写法啊啊啊！
func InitQueue(params map[string]string) (QueueClass) {
	var queue QueueClass
	if params["queue"] == "redis" { // 
		queue = &redis_queue.Queue{
			QueueConf: &redis_queue.QueueConf{

						},
			Pool:		nil,
			Callback:	Callback,
		}
	} else { // list
		queue = &list_queue.Queue{
			Status: 	0,
			Timeout:	-1,
			Callback:	Callback,
		}
	}
	queue.NewQueue()
	return queue
} 