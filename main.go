package main

import (
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

func main() {
	go server()
	time.Sleep(time.Second * 3)
	go client()

	for {
		time.Sleep(time.Second)
	}
}

// 读取消息
func handleConnection(udpConn *net.UDPConn) {

	// 读取数据
	buf := make([]byte, 1024)
	len, udpAddr, err := udpConn.ReadFromUDP(buf)
	if err != nil {
		return
	}
	logContent := strings.Replace(string(buf), "\n", "", 1)
	fmt.Println("server read len:", len)
	fmt.Println("server read data:", logContent)

	// 发送数据
	len, err = udpConn.WriteToUDP([]byte("ok\r\n"), udpAddr)
	if err != nil {
		return
	}

	log.Print("server write len:", len, "\n")
}

// udp 服务端
func server() {

	udpAddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:9998")

	//监听端口
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer udpConn.Close()

	log.Print("udp listening ... ")

	//udp不需要Accept
	for {
		handleConnection(udpConn)
	}
}

func client() {

	udpAddr, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:9998")

	//连接udpAddr，返回 udpConn
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("udp dial ok ")

	// 发送数据
	len, err := udpConn.Write([]byte("hello,wrold!\n\r"))
	if err != nil {
		return
	}
	log.Print("client write len:", len)

	//读取数据
	buf := make([]byte, 1024)
	len, _ = udpConn.Read(buf)
	log.Print("client read len:", len)
	log.Print("client read data:", string(buf))

}
