package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

//连接结构体,很像一个对象
type connection struct {
	ws   *websocket.Conn
	sc   chan []byte //接收的消息？
	data *Data       //连接里的数据对象
}

//设置websocket的读写内存大小
var wu = &websocket.Upgrader{
	ReadBufferSize:  512,
	WriteBufferSize: 512,
	CheckOrigin: func(r *http.Request) bool {
		return true
	}}

func myws(w http.ResponseWriter, r *http.Request) {
	//ws 是一个websocket连接,Upgrade升级http连接为websocket连接
	ws, err := wu.Upgrade(w, r, nil)
	//如果有报错
	if err != nil {
		fmt.Println(err)
		return
	}
	//引用connection结构体并赋值
	c := &connection{sc: make(chan []byte, 256), ws: ws, data: &Data{}}
	//放到h结构体里r元素的管道中
	h.r <- c
	//开一个协程发送消息
	go c.writer()
	c.reader()
	//延迟方法
	defer func() {
		c.data.Type = "logout"
		//减少在线用户
		user_list = del(user_list, c.data.User)
		c.data.UserList = user_list
		c.data.Content = c.data.User
		data_b, _ := json.Marshal(c.data)
		//把用户信息怼到管道里
		h.b <- data_b
		h.r <- c
	}()
}

//这是一个发送消息的方法？
func (c *connection) writer() {
	for message := range c.sc {
		c.ws.WriteMessage(websocket.TextMessage, message)
	}
	c.ws.Close()
}

var user_list = []string{}

//接收消息方法
func (c *connection) reader() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			h.r <- c
			break
		}
		//解析json数据
		json.Unmarshal(message, &c.data)
		switch c.data.Type {
		case "login":
			c.data.User = c.data.Content
			c.data.From = c.data.User
			//用户列表
			user_list = append(user_list, c.data.User)
			c.data.UserList = user_list
			//将数据编码成json字符串
			data_b, _ := json.Marshal(c.data)
			h.b <- data_b
		case "user":
			c.data.Type = "user"
			data_b, _ := json.Marshal(c.data)
			h.b <- data_b
		case "logout":
			c.data.Type = "logout"
			user_list = del(user_list, c.data.User)
			data_b, _ := json.Marshal(c.data)
			h.b <- data_b
			h.r <- c
		default:
			fmt.Print("=====default======")
		}
	}
}

//返回字符串切片
func del(slice []string, user string) []string {
	count := len(slice)
	//
	if count == 0 {
		return slice
	}
	//
	if count == 1 && slice[0] == user {
		return []string{}
	}
	//
	var n_slice = []string{}
	//
	for i := range slice {
		if slice[i] == user && i == count {
			return slice[:count]
		} else if slice[i] == user {
			n_slice = append(slice[:i], slice[i+1:]...)
			break
		}
	}
	fmt.Println("n_slice", n_slice)
	return n_slice
}
