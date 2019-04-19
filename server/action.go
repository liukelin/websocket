package server 

import (
	"time"
	"fmt"
	"io/ioutil"
	"net/http"
)

/**
* [indexAction 请求业务处理]
* @param  {[type]} w http.ResponseWriter [description]
* @param  {[type]} r *http.Request       [description]
* @return {[type]}   [description]
*/
func indexAction(w http.ResponseWriter, r *http.Request) {
	// 解析参数, 默认是不会解析的
	r.ParseForm()

	// d := r.Form["d"]
	// d := r.FormValue("d")
	fmt.Println(time.Now(), "body:", r.Form)
	fmt.Fprintf(w, "success.")
	
	result, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "{\"msg\":\"Body error.\"}")
	}
	body := string(result[:])
	fmt.Fprintf(w, body)
}

/**
* [sendAction 消息发送]
* @param  {[type]} w http.ResponseWriter [description]
* @param  {[type]} r *http.Request       [description]
* @return {[type]}   [description]
*/
func (notice *HandlerNotice) sendAction(w http.ResponseWriter, r *http.Request){
	r.ParseForm()
	
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域

	result, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "{\"msg\":\"Body error.\"}")
	}

	body := string(result[:])
	notice.Queue.Pub(body) // 内容发送到队列

	fmt.Fprintf(w, "{\"msg\":\"success.\"}")
}

/**
* [statusAction 状态查看]
* @param  {[type]} w http.ResponseWriter [description]
* @param  {[type]} r *http.Request       [description]
* @return {[type]}   [description]
*/
func (notice *HandlerNotice) statusAction(w http.ResponseWriter, r *http.Request){
	r.ParseForm()
	
	w.Header().Set("Access-Control-Allow-Origin", "*") //允许访问所有域

	result, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Fprintf(w, "{\"msg\":\"Body error.\"}")
	}

	body := string(result[:])

	fmt.Fprintf(w, "{\"msg\":\"success.\"}"+body+"")
}

