package redis_queue

import (
	"log"
	"fmt"
    "time"
    "context"
	"github.com/garyburd/redigo/redis"
)
/*
  使用发布/订阅类型，使所有节点能够订阅到消息
*/

type QueueClass interface {
	// 连接
	NewQueue()
	// push
	Pub(string) (error)
	// pop
	Sub() (error)
}

type Queue struct {
	QueueConf 	*QueueConf
	Pool		*redis.Pool
	Callback 	func(string) (error)
}

type QueueConf struct {
	Servers []string	 // ["127.0.0.1:7711", "127.0.0.1:7712"]
	Qname	string		 // 队列name
}

func (queue *Queue) NewQueue() {
	pool := &redis.Pool{
        MaxIdle:  10,
        IdleTimeout: 300 * time.Second,
        Dial: func() (redis.Conn, error) {
            c, err := redis.Dial("tcp", queue.QueueConf.Servers[0])
            if err != nil {
                return nil, err
            }
            return c, nil
        },
        TestOnBorrow: func(c redis.Conn, t time.Time) error {
            if time.Since(t) < time.Minute {
                return nil
            }
            _, err := c.Do("PING")
            return err
        },
    }
    log.Printf("new redis pool at: %s", queue.QueueConf.Servers)
    queue.Pool = pool
}

// 发布消息
func (queue *Queue) Pub(message string) (error) {
	
	c := queue.Pool.Get()
    defer c.Close()
    _, err := redis.Int(c.Do("PUBLISH", queue.QueueConf.Qname, message))
    if err != nil {
        return fmt.Errorf("redis publish %s %s, err: %v", queue.QueueConf.Qname, message, err)
    }
	return nil
}

// 订阅消息
func (queue *Queue) Sub() (error) {

    var ctx context.Context
	psc := redis.PubSubConn{Conn: queue.Pool.Get()}
    
    log.Printf("redis pubsub subscribe channel: %v", queue.QueueConf.Qname)
    if err := psc.Subscribe(redis.Args{}.AddFlat(queue.QueueConf.Qname)...); err != nil {
        return err
    }
    done := make(chan error, 1)
    // start a new goroutine to receive message
    go func() {
		defer psc.Close()
        for {
            switch msg := psc.Receive().(type) {
            case error:
                done <- fmt.Errorf("redis pubsub receive err: %v", msg)
                return
			case redis.Message:
                if err := queue.Callback(string(msg.Data[:])); err != nil {
                    log.Printf("queue.Callback: %s", err)
                    // done <- err
                    // return
                }
            case redis.Subscription:
                if msg.Count == 0 {
                    // all channels are unsubscribed
                    done <- nil
                    return
                }
            }
        }
    }()
 
    // health check
    tick := time.NewTicker(time.Minute)
    defer tick.Stop()
    for {
        select {
        case <-ctx.Done():
            if err := psc.Unsubscribe(); err != nil {
                log.Printf("redis pubsub unsubscribe err: %s", err)
                return fmt.Errorf("redis pubsub unsubscribe err: %v", err)
            }
            return nil
        case err := <-done:
            log.Printf("queue.Callback: %s", err)
            return err
        case <-tick.C:
            if err := psc.Ping(""); err != nil {
                log.Printf("redis Ping err: %s", err)
                return err
            }
        }
    }
}

