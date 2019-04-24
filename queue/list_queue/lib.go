package list_queue

/**
  1、使用自带的container/list包 实现本地队列（此为非线程安全，这里使用加锁）
  2、未提供清理功能，长期堆积不消费可能会内存爆炸，这里简单的限制单个list len数量
  3、附带一个自动清理功能（超过过期时间的元素自动清理，因为是非线程安全这个先不考虑用异步回收器做）
  必要实现
  func NewQueue
  func Push
  func Pop
  func Len
  func Empty
*/
import (
    "container/list"
    "sync"
    "time"
)
var lock sync.Mutex

//队列类实现方法
type QueueClass interface {
    // 连接
    NewQueue()
    // publish
    Pub(string) (error)
    // Subscribe
    Sub() (error)
}

type Queue struct {
    List        *list.List  // object
    Status      int       // list 状态，用于限制push 0正常 -1不允许push
    Timeout     int      // 队列元素过期时间 /秒, -1不过期
    Callback 	func(string) (error) // 获取消息后的回调函数
}

// 初始化
func (q *Queue) NewQueue(){
    // q := new(Queue)
    q.List = list.New()
}

// push
func (q *Queue) Pub(message string) error {
    defer lock.Unlock()
    lock.Lock()
    q.List.PushFront(message)
    return nil
}

// sub
func (q *Queue) Sub() (error) {
    for {
        lock.Lock()
        e := q.List.Back()
        if e != nil {
            q.List.Remove(e)
            err := q.Callback(e.Value.(string))
            if err != nil {

            }
        }
        lock.Unlock()

        if e == nil {
            time.Sleep(time.Second)
        }
    }
}

// pop
func (q *Queue) Pop() (interface{}) {
    // 没有直接的pop操作，这两步操作非原子性，因为本类是非线程安全，这里加锁
    defer lock.Unlock()
    lock.Lock()
    e := q.List.Back()
    if e != nil {
        q.List.Remove(e)
        return e.Value
    }
    return nil
}

// len
func (q *Queue) Len() int {
    return q.List.Len()
}

// check len
func (q *Queue) Empty() bool {
    return q.List.Len() == 0
}
