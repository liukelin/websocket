/*
* http server
*/
package server

import (
	"log"
	"net/http"
	"time"
)

// type HandlerNotice *lib.HandlerNotice // 继承lib的HandlerNotice类型
// 用于 goroutine 间通信, 解耦queue
type HandlerNotice struct {
	Push	chan string 
	Pop		chan string
	Status	chan string
	Queue 	QueueClass
}


func Run(params map[string]string){
	HttpServer(params)
	log.Println("Runing...")
}

// 初始化
func InitHandlerNotice(params map[string]string) (*HandlerNotice) {
	return &HandlerNotice{
		Push: make(chan string),
		Pop:  make(chan string),
		Queue: InitQueue(params),
	}
}

/**
* web server 入口
* @return {[type]} [description]
*/
func HttpServer(params map[string]string) {

	notice := InitHandlerNotice(params)

	// http	
	http.HandleFunc("/", indexAction)
	// 发送消息
	http.HandleFunc("/send", notice.sendAction)

	// web socket 服务
	http.Handle("/ws", CreateWebSocketHandler(notice))

	log.Println("http Server Port:", params["port"])

	srv := &http.Server{
		Addr:	":" + params["port"],
		ReadTimeout: 5 * time.Second,	// 设置超时时间，从连接被接受(accept)到request body完全被读取
		WriteTimeout: 10 * time.Second,	// 从request header的读取结束开始，到 response write结束为止 
	}
	err := srv.ListenAndServe()

	// mux := http.NewServeMux()
	// err := http.ListenAndServe(portStr, mux)
	// err := http.ListenAndServe(":" + params["port"], nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}