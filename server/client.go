package server

import (
	"bytes"
	"fmt"
	"log"

	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gorilla/websocket"
)

const (
	//一些限定
	writeWait = 10 * time.Second
	pongWait = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	Ws *Websocket
	//socket连接
	conn *websocket.Conn
	//发送chan
	send chan []byte
	//用户名
	username []byte
	//房间号
	roomID []byte
}

func (c *Client) read() {
	defer func() {
		c.Ws.unregister <- c
		c.conn.Close()
	}()
	//连接限定
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		//消息格式 房间号&用户名:消息
		message = []byte(string(c.roomID) + "&" + string(c.username) + ":" + string(message))
		//fmt.Println(string(message))
		c.Ws.broadcast <- message
	}
}

func (c *Client) write() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				log.Printf("error: %v", err)
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

//聊天室请求
type ChatRequest struct {
	RId    string `json:"r_Id" form:"room_id"`
	UName  string `json:"u_Name" form:"user_name"`
}

//socket服务和协议升级
func ServeWs(ws *Websocket, c *gin.Context) {
	// 获取前端数据
	var req ChatRequest
	if err := c.ShouldBind(&req); err != nil {
		log.Printf("ServeWs err:%v\n", err)
		c.JSON(http.StatusOK, gin.H{"errno": "-1", "errmsg": "参数不匹配，请重试"})
		return
	}
	userName := c.Query("user_name")
	roomID := c.Query("room_id")
	fmt.Println("user_name:" + userName + " room_id:" + roomID)
	//将网络请求升级为websocket
	var upgrader = websocket.Upgrader{
		// 解决跨域问题
		CheckOrigin: func(r *http.Request) bool {
			//fmt.Println("1)check!")
			return true
		},
	}
	//fmt.Println("2)check!")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println(err)
		return
	}
	//fmt.Println("3)check!")
	fmt.Println(userName)
	client := &Client{Ws: ws, conn: conn, send: make(chan []byte, 256), username: []byte(userName), roomID: []byte(roomID)}
	client.Ws.register <- client

	go client.write()
	go client.read()
}
