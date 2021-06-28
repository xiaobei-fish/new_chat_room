package server

import (
	"fmt"
	"strings"
)

type Websocket struct {
	//登录的客户端
	clients map[*Client]bool
	//广播chan
	broadcast chan []byte
	//登录请求chan
	register chan *Client
	//退出登录chan
	unregister chan *Client
	//房间号字典 key:client value:房间号
	roomID map[*Client]string
}

func NewWs() *Websocket {
	return &Websocket{
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		roomID:     make(map[*Client]string),
	}
}
//运行监听
func (ws *Websocket) Run() {
	for {
		select {
		case client := <-ws.register:
			ws.clients[client] = true                 // 注册client端
			ws.roomID[client] = string(client.roomID) // 给client端添加房间号
			for c := range ws.clients {
				if string(c.roomID) == string(client.roomID) {
					msg := string(client.username) + "已加入聊天室"
					fmt.Println(msg)
					c.send <- []byte(msg)
				}
			}
		case client := <-ws.unregister:
			for c := range ws.clients {
				if string(c.roomID) == string(client.roomID) {
					msg := string(client.username) + "已退出聊天室"
					fmt.Println(msg)
					c.send <- []byte(msg)
				}
			}
			if _, ok := ws.clients[client]; ok {
				delete(ws.clients, client)
				delete(ws.roomID, client)
				close(client.send)
			}
		case message := <-ws.broadcast:
			for client := range ws.clients {
				//消息格式 房间号&用户名:消息
				//消息切割得到房间号，并向房间内所有用户广播
				msg := strings.Split(string(message), "&")
				if string(client.roomID) == msg[0] {
					select {
					case client.send <- []byte(msg[1]):
					default:
						close(client.send)
						delete(ws.clients, client)
						delete(ws.roomID, client)
					}
				}
			}
		}
	}
}
