# golang编写的websocket服务端

## 简介

高可用下
sendMsg -> http server -> pub/sub
-> websocket server -> sendMsg

单机下
sendMsg -> http server -> list
-> websocket server -> sendMsg

其他建议
作为消息不落地，建议使用单机环境。 但解决高可用问题的话 配置 upstream backup，当前只提供一个节点在服务，其余的睡眠
upstream  wsserver  {
        server   127.0.0.1:8080;
        server   127.0.0.2:8080 backup;
}


```
    1、queue的发布订阅/本地list 作为消息转介

    2、http server作为消息发送入口

    3.创建发送消息格式：json   
                    {
                        "topic":uids/all ,             // 发送对象 uids:指定多个用户，all:发给所有在线用户
                        "topic_ids": ["1",2"...],      // 当topic为uids，topic_ids为uid列表。 
                        "types":"notice"               // 消息类型，用于订阅不同消息 默认为 notice
                        "msg":"消息内容" ,              // 消息内容，透传参数，具体格式 具体业务进行讨论对接。
                        "timeout":1212323,	   		   // 消息过期时间戳
                        "time":"2019-12-12 12:12:12"   //消息发送时间"     
                    }

    4.websocket 推送返回格式：json              
                    {
                        "types":"notice"               消息类型，用于订阅不同消息
                        "msg":"消息内容" ,              json 具体格式 具体业务进行讨论对接。
                        "time":"2019-12-12 12:12:12"   //消息发送时间" 
                    }
```

## 编译

```sh

# windows
./build.sh windows

# linux
./build.sh linux

#mac
./build.sh darwin

#freebsd
./build.sh freebsd

```

## nginx 配置
```

upstream websocket_proxy {
    server  127.0.0.1:3210;
    server  127.0.0.1:3211 backup;
}

server {
    listen       3220;
    server_name  _;
    
    location ~ ^/ws$ {
        proxy_pass      http://websocket_proxy;
        proxy_set_header Host $host:$server_port;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    location / {
        proxy_pass      http://websocket_proxy;
        # include         proxy_params;
    }
}

```

## 启动服务
go run main.go -post=3210 -queue=list
                                redis/localhost:6379
                                disque/localhost:7711


## web demo页面
./demo/socket.html


## 压测
$ ab -c 100 -n 100 -T 'application/json' -p ./postdata.txt http://localhost:3210/send
$ webbench -c 100 -t 60 -F "./postdata.txt" http://localhost:3210/send
