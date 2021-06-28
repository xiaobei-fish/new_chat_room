package main

import (
	"NewTest4/server"
	_ "NewTest4/server"
	"errors"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"time"
)
type User struct {
	Name 	string
	RoomId 	string
}

var roomId string
var username string

func main() {
	fmt.Println("Starting application...")
	r := gin.New()
	r.Use(gin.Recovery())
	r.LoadHTMLFiles("./views/login.html","./views/chat.html")
	r.Static("static/", "./static")

	r.GET("/login",func(c *gin.Context){
		c.HTML(200, "login.html", nil)
	})
	r.POST("/login",func(c *gin.Context){
		var resUser User
		_ = c.ShouldBind(&resUser)
		ip, err := externalIP()
		if err != nil{
			fmt.Println(err)
		}
		name := resUser.Name + "(ip:" + ip.String() + ")"
		room := resUser.RoomId
		//fmt.Println(name + " " + room)
		roomId = room
		username = name

		Success(c, gin.H{"username": name, "room_num": roomId}, "登录成功", roomId, username)
	})
	r.GET("/chat_room",func(c *gin.Context){
		c.HTML(http.StatusOK,"chat.html", gin.H{
			"user_name": username,
			"room_id"  : roomId,
		})
	})
	//设置跨域
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"x-xq5-jwt", "Content-Type", "Origin", "Content-Length"},
		ExposeHeaders:    []string{"x-xq5-jwt"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	Ws := server.NewWs()
	go Ws.Run()

	r.GET("/ws",func(c *gin.Context){
		server.ServeWs(Ws, c)
	})

	_ = r.Run(":8080")
}
//响应体封装
func Response(ctx *gin.Context,httpStatus int,code int,data gin.H,msg string,rId string,uName string){
	ctx.JSON(httpStatus,gin.H{
		"code":code,
		"data":data,
		"msg" :msg,
		"r_Id" :rId,
		"u_Name":uName,
	})
}
//成功返回JSON
func Success(ctx *gin.Context,data gin.H,msg string,rId string,uName string){
	Response(ctx,http.StatusOK,http.StatusMovedPermanently,data,msg,rId,uName)
}
//取Ip
func externalIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			ip := getIpFromAddr(addr)
			if ip == nil {
				continue
			}
			return ip, nil
		}
	}
	return nil, errors.New("connected to the network?")
}
func getIpFromAddr(addr net.Addr) net.IP {
	var ip net.IP
	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	}
	if ip == nil || ip.IsLoopback() {
		return nil
	}
	ip = ip.To4()
	if ip == nil {
		return nil // not an ipv4 address
	}
	return ip
}