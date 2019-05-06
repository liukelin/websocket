package server

import (
    "fmt"
    "log"
    "golang.org/x/net/websocket"
    "net/http"
    "time"
    "websocket/lib"
)

// websocket opcode 
// TextMessage
const TextMessage = 1
// BinaryMessage denotes a binary data message.
const BinaryMessage = 2
// CloseMessage
const CloseMessage = 8
// PingMessage 0x9
const PingMessage = 9
// PongMessage 0xA
const PongMessage = 10


// 连接object
type Ws struct {
    cl      ClientList
    notice  *HandlerNotice    // Queue       
}

// 存储 websocket.Conn 与 用户关系, map[uid][uuid]
// 允许同一个uid建立多个连接, 后续需要限制个数
type ClientList map[string]map[string]*Client

// 连接客户端信息
type Client struct {
    Uuid            string              // 每个连接的唯一标识
    Conn            *websocket.Conn     // 连接对象
    ClientParams    *lib.ClientParams   // url 连接参数
    LocalAddr       string              // 冗余url
}

// 创建ws服务
func CreateWebSocketHandler(notice *HandlerNotice) (http.Handler) {

    // 客户端连接池
    ws := &Ws{
        cl:     ClientList{}, 
        notice: notice,
    }
    // 监听新消息
    ws.ListeningMsg()

    ht := websocket.Handler(ws.wsHandler)
    // var w http.ResponseWriter
    // var r *http.Request
    // w.Header().Set("Access-Control-Allow-Origin", "*") 
    // r.Header.Del("Origin")
    // ht.ServeHTTP(w, r)
    // 启动websocket服务
    return ht
}


// 连接主逻辑func goroutine
// 默认不包含ping pong
func (ws *Ws) wsHandler(conn *websocket.Conn) {
    
    var uid string
    var uuid string
    request := make([]byte, 128);
    defer func(){
        // gc
        request = nil
        ws.delClientList(uid, uuid)
        conn.Close(); // 连接关闭
    }()

    // 连接认证
    clientParams := lib.ClientUrlParamsParse(conn.LocalAddr().String())
    if clientParams == nil {
        return
    }

    // 保存连接信息
    uid = clientParams.Uid
    uuid = lib.GetUuid()
    ws.addClientList(&Client{
        Uuid:  uuid,
        Conn:  conn,
        ClientParams: clientParams,
        LocalAddr: conn.LocalAddr().String(),
    })

    log.Println("Client connection:", conn.LocalAddr().String())

    // 自行实现ping/pong
    go func(){
        for {
            // ping
            conn.PayloadType = byte(PingMessage)
            conn.Write([]byte{0})
            time.Sleep(2 * time.Second)
        }
    }()

    // 监听客户端发送的消息
    for {
        // 获取客户端发送的消息 ，此函数为阻塞函数
        readLen, err := conn.Read(request)
        if lib.CheckErr(err, "Client connection close: "+conn.LocalAddr().String()) {
            break
        }

        // 获取消息类型
        if readLen == 0 {
            // socket被关闭
            log.Println("Client connection close: ",conn.LocalAddr().String())
            break;
        } else {
            fmt.Println(string(conn.PayloadType))
            //输出接收到的信息
            fmt.Println(string(request[:readLen]))
            time.Sleep(time.Second)
            //发送
            conn.PayloadType = byte(TextMessage)
            conn.Write([]byte("World !"))
        }
        request = make([]byte, 128)
    }
}


// 保存客户连接信息
func (ws *Ws) addClientList(client *Client){
    if _, ok := ws.cl[client.ClientParams.Uid]; ok {
        ws.cl[client.ClientParams.Uid][client.Uuid] = client
    }else{
        ws.cl[client.ClientParams.Uid] = map[string]*Client{client.Uuid:client}
    }
}

// 删除客户连接信息
func (ws *Ws) delClientList(uid, uuid string){
    if _, ok := ws.cl[uid]; ok {
        delete(ws.cl[uid], uuid)// 踢出用户
    }
}


// 发送消息
func (cl ClientList) SendMsg(msg *lib.SendMsgContent){
    var clients []*lib.ClientParams // 这个可以缓存起来，不需要每次创建

    // 取出客户端
    for _, v := range cl {
        for _, t := range v {
            clients = append(clients, t.ClientParams)
        }
    }

    // 解析出接收客户端
    uids := lib.GetAllowUids(clients, msg)

    // 转换发送格式
    echoMsg := lib.SendToEchoMsgJson(msg)

    // 发送
    for _,v := range uids {
        if _, ok := cl[v]; !ok {
            continue
        }
        for _, client := range cl[v] {
            client.Conn.PayloadType = byte(TextMessage)
            client.Conn.Write(echoMsg)
        }
    }
}

// 监听接收接口消息
func  (ws *Ws) ListeningMsg(){
    // 优化点1：go routine 这里可以开多个goroutine 并发消费
    // 优化点2：此处可使用 select channel 阻塞监听，而不是轮询list pop
    go func(ws *Ws){
        err := ws.notice.Queue.Sub()
        if err != nil {
            log.Println("Queue subscribe err: ", err)
        }
    }(ws)
}
