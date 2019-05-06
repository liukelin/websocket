package lib

import (
	"strconv"
	"log"
	"time"
	"encoding/json"
	"net/url"
	"websocket/queue/list_queue"
)

const TOPIC_ALL = "all"
const TOPIC_UIDS = "uids"
const DEFAULT_TYPE = "notice"

// 用于 goroutine 间通信, 解耦queue
type HandlerNotice struct {
	Push	chan string 
	Pop	chan string
	Queue 	*list_queue.Queue
}

// 连接url必要参数, 代表每一个连接客户端
// ?d={"uid":"1","types":["notice"]}
type ClientParams struct {
    Types []string   // 订阅的消息类型（可订阅多种类型）
    Uid   string     // 客户端唯一标识
}

// 消息接收格式 json
// {"topic":"all","topic_ids":["1"],"types":"notice","msg":"测试消息","timeout":2121325421,"time":"1999-12-12"}
type SendMsgContent struct {
	Topic		string 	 `json:"topic"`		  // 接收类型 uids/all
	Topic_ids	[]string `json:"topic_ids"`	  // 接收用户
	Types		string	 `json:"types"`		  // 订阅的消息类型, 目前规定一条消息的类型为固定，不能属于多个
	Msg		string	 `json:"msg"`		  // 消息内容
	Timeout		int	 `json:"timeout"`	  // 消息过期时间戳
	Time		string	 `json:"time"`		  // 消息发送时间
}

// 返回给客户端格式：json 
// {"types":"notice", "msg":"消息内容", "time":"消息时间" }
type EchoMsgContent struct {
	Types		string	 `json:"types"`		// 目前规定一条消息的类型为固定，不能属于多个
	Msg		string	 `json:"msg"`
	Time		string	 `json:"time"`
}


// 统一入队、出队函数
// 常驻
func QueuePushPop(notice *HandlerNotice){
	go func(){
        select {
		case msg := <-notice.Push: // 接收接口发送消息
            		log.Printf("push:%s \r\n", msg)
        }
    }()
}

// 生成唯一标识
// 后期可改成正式的uuid生成
func GetUuid() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

// 将接收到的json格式消息转换成 SendMsgContent
func JsonMsgToSendMsg(msg string) (*SendMsgContent) {

	var sendMsg SendMsgContent
	err := json.Unmarshal([]byte(msg), &sendMsg)
	if err != nil {
		log.Println("JsonMsgToSendMsg json.Unmarshal:", err)
		return nil
	}

	// 检测格式合法
	// ...
	return &sendMsg
}

// 将发送消息转换成json格式返回消息
func SendToEchoMsgJson(msg *SendMsgContent) []byte {
	echoMsg := &EchoMsgContent{
		Types: 	msg.Types,
		Msg: 	msg.Msg,
		Time:	msg.Time,
	}
	data, err := json.Marshal(echoMsg)
    	if err != nil {
		log.Println("Json.Marshal failed:", err)
		return nil
	}
	// 必要参数检查
	// ...

	return data
}
 
// 解析url格式参数
func UrlParse(uri string) (url.Values){
	u, err := url.Parse(uri)
	if err != nil {
		log.Println("UrlParse:", err)
		return nil
	}
	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		log.Println("Url.ParseQuery:", err)
		return nil
	}
	return values
}

// 整理连接参数返回
func ClientUrlParamsParse(url string) (*ClientParams) {
	uri := UrlParse(url)

	defer func(){
		uri = nil
    	}()

	if _, ok := uri["d"]; !ok {
		log.Println("url d null.")
		return nil
	}

	var clientParams ClientParams
	err := json.Unmarshal([]byte(uri["d"][0]), &clientParams)
	if err != nil {
		log.Println("json.Unmarshal:", err)
		return nil
	}
	// 验证必要参数是否都已经存在
	// ...

	return &clientParams
}

// 错误处理函数
func CheckErr(err error, extra string) bool {
    if err != nil {
        formatStr := " Err : %s";
        if extra != "" {
            formatStr = extra + formatStr;
        }
	log.Println(formatStr, err.Error())
        return true;
    }
    return false;
}

// 检查数组是否存在某元素
func InArray(arr []string, value string) bool {
	for _,v := range arr {
		if value == v {
			return true
		}
	}
	return false
}

// 筛选出需要接收消息的客户端
// return [uid...]
func GetAllowUids(cliens []*ClientParams, msg *SendMsgContent) ([]string){
	repeat := make(map[string]int, 0) // 避免对list去重操作影响性能，使用map先记录
	
	defer func(){
		repeat = nil
	}()
	
	var uids []string
	for _, v := range cliens {

		// 检查定于的频道是否包含
		if !InArray(v.Types, msg.Types) {
			continue
		}

		// 检查发送目标是否满足
		if msg.Topic == TOPIC_UIDS { // 针对用户群体
			if InArray(msg.Topic_ids, v.Uid) {
				repeat[v.Uid] = 0
			}
		} else if msg.Topic == TOPIC_ALL { // 全部群发
			repeat[v.Uid] = 0
		} else {
			continue
		}
	}
	// 去重
	for k, _ := range repeat {
		uids = append(uids, k)
	}
	return uids
}
