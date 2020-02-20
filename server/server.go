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
	addressX *net.UDPAddr

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
	s.addressX, err = net.ResolveUDPAddr("udp4", "127.0.0.1:1541")
	if err != nil {
		return
	}

	s.addressA, err = net.ResolveUDPAddr("udp4", "127.0.0.1:9737")
	if err != nil {
		return
	}

	s.addressB, err = net.ResolveUDPAddr("udp4", "127.0.0.1:1542")
	if err != nil {
		return
	}

	s.socketA, err = net.DialUDP("udp", s.addressX, s.addressA)
	if err != nil {
		return
	}

	//DONT USE LISTEN!

	s.socketB, err = net.DialUDP("udp", nil, s.addressB)
	if err != nil {
		return
	}
	//ticker := time.NewTicker(s.rate * time.Millisecond)

	s.stop = make(chan bool)
	go func() {
		for {
			select {
			case <-s.stop:
				return
			default:
				s.update()
			}
		}
	}()

	s.receive()
}
func (s *Server) Stop() {
	s.stop <- true
}

//func (s *Server) update(c chan bool) {
func (s *Server) update() {
	start := time.Now()
	end := time.Now().Add(s.rate * time.Millisecond)

	err := s.socketA.SetReadDeadline(time.Now().Add((s.rate) * time.Millisecond))
	if err != nil {
		return
	}

	data := []byte{0}

	// write to ask for context
	s.socketA.Write(data)

	// wait for context
	buffer := make([]byte, 1024)
	_, err = s.socketA.Read(buffer)
	if err != nil {
		fmt.Printf("%v\n", err)
		fmt.Printf("tick: %v (%vms) (timeout)\n", s.tick, time.Now().Sub(start).Milliseconds())
		return
	}

	// unmarshal context
	//packet := &gamedata.Packet{}
	//err = proto.Unmarshal(buffer[:l], packet)
	//if err != nil {
	//	log.Println(err)
	//}

	fmt.Printf("tick: %v (%vms trip) (%vms sleep)\n", s.tick, time.Now().Sub(start).Milliseconds(), end.Sub(time.Now()).Milliseconds())
	s.tick++
	time.Sleep(end.Sub(time.Now()))
}

func (s *Server) receive() {
	fmt.Printf("receiving\n")
	for {
		// wait for context
		buffer := make([]byte, 1024)
		l, err := s.socketB.Read(buffer)
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
		fmt.Println("here")
	}
}
