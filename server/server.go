package server

import (
	"fmt"
	"log"
	"net"
	"time"

	//"github.com/fanmanpro/coordinator-server/gamedata"

	"github.com/fanmanpro/game-server/serializable"
	"github.com/golang/protobuf/proto"
)

type Server struct {
	localSimAddress  *net.UDPAddr
	remoteSimAddress *net.UDPAddr

	localClientAddress  *net.UDPAddr
	remoteClientAddress *net.UDPAddr

	socketSim    *net.UDPConn
	socketClient *net.UDPConn

	rate   time.Duration
	tick   int32
	seater Seater

	stop chan bool
}

func NewServer() *Server {
	return &Server{rate: 50, tick: 10} // 100 is 10 ticks per second, 50 is 20, 33 is 30, etc.
}

func (s *Server) Start() error {
	var err error
	s.localSimAddress, err = net.ResolveUDPAddr("udp4", "127.0.0.1:1541")
	if err != nil {
		return err
	}

	s.remoteSimAddress, err = net.ResolveUDPAddr("udp4", "127.0.0.1:9737")
	if err != nil {
		return err
	}

	s.localClientAddress, err = net.ResolveUDPAddr("udp4", "127.0.0.1:1542")
	if err != nil {
		return err
	}

	s.remoteClientAddress, err = net.ResolveUDPAddr("udp4", "127.0.0.1:9738")
	if err != nil {
		return err
	}

	s.socketSim, err = net.DialUDP("udp", s.localSimAddress, s.remoteSimAddress)
	if err != nil {
		return err
	}
	//go s.receiveSim()

	s.socketClient, err = net.DialUDP("udp", s.localClientAddress, s.remoteClientAddress)
	if err != nil {
		return err
	}

	// start the server ticker. careful for this one.
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

	s.receiveClient()
	return nil
}
func (s *Server) Stop() {
	s.stop <- true
}

func (s *Server) broadcast(data []byte) {
	//err := s.socketClient.SetWriteDeadline(time.Now().Add((s.rate) * time.Millisecond))
	//if err != nil {
	//	fmt.Printf("%v\n", err)
	//	return
	//}
	s.socketClient.Write(data)
}

var sent time.Time
var end time.Time

func (s *Server) update() {
	end = time.Now().Add(s.rate * time.Millisecond)

	// set the read deadline for when we are waiting for the returned tick
	err := s.socketSim.SetReadDeadline(end)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	// wait for a context to come back from the sim
	buffer := make([]byte, 1024)
	l, err := s.socketSim.Read(buffer)
	if err != nil {
		//fmt.Printf("%v\n", err)
		return
	}

	// unmarshal the incoming context
	incoming := &serializable.Context3D{}
	readBuffer := buffer[0:l]
	err = proto.Unmarshal(readBuffer, incoming)
	if err != nil {
		fmt.Printf("%v\n", err)
	}

	// check if the incoming tick is the same as the current expected tick
	if incoming.Tick != s.tick {
		fmt.Printf("context was dropped: %v vs %v (old)\n", incoming.Tick, s.tick)
		//return
	} else {
		go s.broadcast(readBuffer)
	}

	//fmt.Printf("%+v\n", len(incoming.Transforms))

	//data, err = proto.Marshal(incoming)
	//if err != nil {
	//	fmt.Printf("%v\n", err)
	//	return
	//}
	//s.socketClient.Write(data)

	//fmt.Printf("tick: %v (%vms trip) (%vms sleep)\n", s.tick, time.Now().Sub(start).Milliseconds(), end.Sub(time.Now()).Milliseconds())

	// define the outgoing packet we want to send to the sim
	s.tick++
	//fmt.Printf("tick incremented\n")

	outgoing := &serializable.Context3D{
		Tick: s.tick,
	}

	// marshal the outgoing packet
	data, err := proto.Marshal(outgoing)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	time.Sleep(end.Sub(time.Now()))

	// write the outgoing packet to the socket shared with the sim
	s.socketSim.Write(data)
}

// received a context from either
// NO GET RID OF THIS. FROM THE SIM WE READ AFTER WE WROTE AS A LOOP. NOT THIS SEPARATE READ LOOP.
//func (s *Server) receiveSim() {
//	fmt.Printf("receiving for sim\n")
//	for {
//		// wait for context
//		buffer := make([]byte, 1024)
//		l, err := s.socketSim.Read(buffer)
//		if err != nil {
//			//fmt.Printf("[UDP] Failed reading %v. Reason: %v\n", addr.String(), err.Error())
//			return
//		}
//
//		// unmarshal context coming from sim. authoritative!
//		context := &serializable.Context{}
//		err = proto.Unmarshal(buffer[:l], context)
//		if err != nil {
//			log.Printf("invalid packet from sim: %v\n", err)
//		}
//
//		// compare
//		fmt.Printf("%+v", context)
//		break
//	}
//}

func (s *Server) receiveClient() {
	fmt.Printf("receiving for client\n")
	for {
		// wait for context
		buffer := make([]byte, 1024)
		l, err := s.socketClient.Read(buffer)
		if err != nil {
			//fmt.Printf("[UDP] Failed reading %v. Reason: %v\n", addr.String(), err.Error())
			return
		}

		// unmarshal context from client. validate then forward.
		context := &serializable.Context3D{}
		err = proto.Unmarshal(buffer[:l], context)
		if err != nil {
			log.Println(err)
		}

		// compare
		fmt.Println("here")
	}
}
