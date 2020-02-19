package server

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/fanmanpro/coordinator-server/gamedata"
	"github.com/golang/protobuf/proto"
)

type Server struct {
	socketA  *net.UDPConn
	addressA *net.UDPAddr

	socketB  *net.UDPConn
	addressB *net.UDPAddr

	rate   time.Duration
	tick   int
	seater Seater

	stop chan bool
}

func NewServer() *Server {
	return &Server{rate: 100} // 100 is 10 ticks per second, 50 is 20, 33 is 30, etc.
}

func (s *Server) Start() {
	var err error
	s.addressA, err = net.ResolveUDPAddr("udp4", "127.0.0.1:1541")
	if err != nil {
		return
	}

	s.addressB, err = net.ResolveUDPAddr("udp4", "127.0.0.1:1542")
	if err != nil {
		return
	}

	s.socketA, err = net.ListenUDP("udp", s.addressA)
	if err != nil {
		return
	}

	DONT USE LISTEN!

	s.socketB, err = net.UDP.ListenUDP("udp", s.addressB)
	if err != nil {
		return
	}

	ticker := time.NewTicker(s.rate * time.Millisecond)
	s.stop = make(chan bool)
	go func() {
		for {
			select {
			case <-s.stop:
				return
			case <-ticker.C:
				s.update()
				s.tick++
			}
		}
	}()

	s.receive()
}
func (s *Server) Stop() {
	s.stop <- true
}

func (s *Server) update() {
	start := time.Now()
	// write to ask for context
	s.socketA.WriteToUDP(make([]byte, 0), s.addressA)

	// wait for context
	buffer := make([]byte, 1024)
	l, _, err := s.socketA.ReadFromUDP(buffer)
	if err != nil {
		//fmt.Printf("[UDP] Failed reading %v. Reason: %v\n", addr.String(), err.Error())
		return
	}

	// unmarshal context
	packet := &gamedata.Packet{}
	err = proto.Unmarshal(buffer[:l], packet)
	if err != nil {
		log.Println(err)
	}

	// compare
	//fmt.Println("here")
	fmt.Printf("duration: %v\n", time.Now().Sub(start).Nanoseconds())
}

func (s *Server) receive() {
	for {
		// wait for context
		buffer := make([]byte, 1024)
		l, addr, err := s.socketB.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("[UDP] Failed reading %v. Reason: %v\n", addr.String(), err.Error())
			return
		}

		// unmarshal context
		packet := &gamedata.Packet{}
		err = proto.Unmarshal(buffer[:l], packet)
		if err != nil {
			log.Println(err)
		}

		// compare
		fmt.Println("here")
	}
}
