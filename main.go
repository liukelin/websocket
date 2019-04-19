package main

import (
	"log"
	"flag"
	"websocket/server"
)


// 参数接收
var Port  *string 	  // TCP监听端口
var Queue *string	  // 使用消息队列类型 redis/disque

func main(){
	
	Port = flag.String("port", "3210", "服务端口端口")
	Queue = flag.String("queue", "list", "使用消息中间类型: (list or redis//localhost:6379)")
	flag.Parse()

	server.Run(map[string]string{
		"port": *Port,
	})
	log.Println("Runing...")
}