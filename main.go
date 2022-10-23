package main

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/pion/rtp/v2"
)

const (
	MTU         = 1500
	PayloadType = 100
	SSRC        = 1234
	clockrate   = 1000
)

func main() {
	LogConfig()

	go server()
	time.Sleep(time.Second * 3)
	go client()

	for {
		time.Sleep(time.Second)
	}
}

// 读取消息
func handleConnection(udpConn *net.UDPConn, f *os.File) {

	buf := make([]byte, MTU+10)
	_, err := udpConn.Read(buf)
	if err != nil {
		log.Fatal(err)
	}
	packet := rtp.Packet{}
	packet.Unmarshal(buf)

	fmt.Printf("sequence number: %d\n\r", packet.Header.SequenceNumber)

	data := packet.Payload

	_, err = f.Write(data)
	if err != nil {
		log.Print(err)
	}
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

	f, err := os.OpenFile("resource/test", os.O_RDWR|os.O_CREATE, 644)
	if err != nil {
		log.Fatal(err)
	}

	//udp不需要Accept
	for {
		handleConnection(udpConn, f)
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

	//发送流数据
	//从数据源读取数据
	fd, err := os.Open("resource/EdgeSite_arch.PNG")
	if err != nil {
		log.Fatal(err)
	}

	dataChann := make(chan []byte, 10)

	go func() {
		//the function read data from source
		for {
			buf := make([]byte, MTU*5)
			index, err := fd.Read(buf)
			if err != nil {
				log.Print(err)
				log.Printf("read length: %d, buffer length: %d\n\r", index, len(buf))
				dataChann <- buf[:index]
				break
			}
			log.Printf("read length: %d, buffer length: %d\n\r", index, len(buf))
			dataChann <- buf
		}
		close(dataChann)
		defer fd.Close()
	}()

	go func() {
		sequencer := 1000
		packetizer := rtp.NewPacketizer(MTU, PayloadType, SSRC, &RRPayloader{}, rtp.NewFixedSequencer(uint16(sequencer)), clockrate)

		for data := range dataChann {
			packets := packetizer.Packetize(data, 1)
			for _, p := range packets {
				rawBuf, err := p.Marshal()
				if err != nil {
					panic(err)
				}
				_, err = udpConn.Write(rawBuf)
				if err != nil {
					log.Print(err)
				}
				sequencer++
			}
		}

		defer udpConn.Close()
	}()

}

type RRPayloader struct {
}

func (rr *RRPayloader) Payload(mtu int, payload []byte) [][]byte {
	reader := bytes.NewReader(payload)

	payloads := make([][]byte, 1)

	for {
		buf := make([]byte, mtu)
		index, err := reader.Read(buf)
		/* the error handle is vital, but i am not skilled */
		if index < mtu {
			payloads = append(payloads, buf[:index])
			log.Printf("read length:%d\n\r", index)
			log.Print(err)
			break
		} else {
			payloads = append(payloads, buf)
		}
	}

	return payloads
}

func LogConfig() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
}
