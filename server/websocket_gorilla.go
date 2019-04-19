package server

import (
	"net/http"
    // "websocket/lib"
	"github.com/gorilla/websocket"
	
)

// http升级websocket协议的配置
var wsUpgrader = websocket.Upgrader{
	// 允许所有跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 读取消息
func (ws *Ws) wsReadLoop() {

}

// func (ws *Ws) wsHandler_gori(resp http.ResponseWriter, req *http.Request) {
// 	// 应答客户端告知升级连接为websocket
// 	wsSocket, err := wsUpgrader.Upgrade(resp, req, nil)
// 	if err != nil {
// 		return
// 	}

// 	var uid string
//     var uuid string
//     request := make([]byte, 128);
//     defer func(){
//         // gc
//         request = nil
//         ws.delClientList(uid, uuid)
//     }()

//     // 连接认证
//     clientParams := lib.ClientUrlParamsParse(conn.LocalAddr().String())
//     if clientParams == nil {
//         return
//     }

// 	// 保存连接信息
//     uid = clientParams.Uid
//     uuid = lib.GetUuid()
//     ws.addClientList(&Client{
//         Uuid:  uuid,
//         Conn:  wsSocket.Conn,
//         ClientParams: clientParams,
//         LocalAddr: conn.LocalAddr().String(),
//     })
// }
