package main

import (
	"encoding/json"
	"fmt"
)

//定一个结构体
var h = hub{
	c: make(map[*connection]bool),
	b: make(chan []byte),
	r: make(chan *connection),
	u: make(chan *connection),
}

//先声明一个结构体
type hub struct {
	c map[*connection]bool
	b chan []byte
	r chan *connection
	u chan *connection
}

//这是什么
func (h *hub) run() {
	for {
		//select类似switch语句，但是select是随机执行一个可运行的case，如果case不能运行就会阻塞
		select {
		case c := <-h.r:
			//这里是登录处理的方法
			fmt.Println("1")
			h.c[c] = true
			c.data.Ip = c.ws.RemoteAddr().String()
			c.data.Type = "handshake"
			c.data.UserList = user_list
			data_b, _ := json.Marshal(c.data)
			c.sc <- data_b
		case c := <-h.u:
			fmt.Println("2")
			if _, ok := h.c[c]; ok {
				delete(h.c, c)
				close(c.sc)
			}
		case data := <-h.b:
			//发送消息处理方法
			fmt.Println("3")
			for c := range h.c {
				select {
				case c.sc <- data:
				default:
					delete(h.c, c)
					close(c.sc)
				}
			}
		}
	}
}
