package udp

import (
	"fmt"
	"log"
	"net"

	"github.com/fanmanpro/coordinator-server/client"
	"github.com/fanmanpro/coordinator-server/gamedata"
	"github.com/golang/protobuf/proto"
)

// Server to open web socket ports
type Server struct {
	ip            string
	port          string
	conn          *net.UDPConn
	acceptedAddrs []*net.UDPAddr
	ReadQueue     chan *gamedata.Packet
	AcceptQueue   chan *client.AcceptUDPPacket
}

// NewServer initializes a new UDP server without starting it
func NewServer(ip string, port string) *Server {
	return &Server{ip: ip, port: port, acceptedAddrs: make([]*net.UDPAddr, 0), ReadQueue: make(chan *gamedata.Packet, 10), AcceptQueue: make(chan *client.AcceptUDPPacket, 2)}
}

// AcceptAddress todo
func (u *Server) AcceptAddress(addr *net.UDPAddr) {
	u.acceptedAddrs = append(u.acceptedAddrs, addr)
}

// Start starts the already intialized UDPServer
func (u *Server) Start() error {
	addr, err := net.ResolveUDPAddr("udp4", u.ip+":"+u.port)
	if err != nil {
		return err
	}

	// we want to be able to tell the write loop when the connection will be close so it can stop writing to it
	//stop := make(chan bool, 1)

	// setup listeners for incoming UDP connection
	// we are treating these as separate UDP sockets because the simulation client's port must be closed to public
	// https://stackoverflow.com/questions/35300039/in-golang-how-to-receive-multicast-packets-with-socket-bound-to-specific-addres
	u.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	//do one of these for each connected client from game-server.go once the running phase starts

	//go u.readConn(u.simConn, stop, u.queuePacketSim)

	//start := make(chan bool, 1)
	go u.procConn(u.conn)

	//go u.writeConn(stop)

	return nil
}

func (u *Server) procConn(c *net.UDPConn) {
	buffer := make([]byte, 1024)

	for {
		unknown := true
		//fmt.Printf("[UDP] Awaiting to read\n")
		l, addr, err := c.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("[UDP] Failed reading %v. Reason: %v\n", addr.String(), err.Error())
			break
		}

		for _, a := range u.acceptedAddrs {
			//fmt.Printf("[UDP] ---------------------------- Accept Addr: %v vs %v == %v\n", a, addr, a.String() == addr.String())
			if a.String() == addr.String() {
				unknown = false
				break
			}
		}
		if unknown {
			u.AcceptAddress(addr)
			//fmt.Printf("[UDP] ---------------------------- Accepting Addr: %v\n", addr)
		}

		//fmt.Printf("[UDP] Reading %v bytes for %v\n", l, addr.String())
		packet := &gamedata.Packet{}
		err = proto.Unmarshal(buffer[:l], packet)
		if err != nil {
			log.Println(err)
		}

		if unknown {
			unknown = false
			send := make(chan *gamedata.Packet, 1)
			go u.writeAddr(c, addr, send)
			acceptPacket := &client.AcceptUDPPacket{
				CID:  packet.Cid,
				Addr: addr,
				Send: send,
			}
			fmt.Printf("[UDP] Reading was accept\n")
			u.AcceptQueue <- acceptPacket
			//fmt.Printf("[UDP] Read accept\n")
		}

		fmt.Printf("%v from %v\n", packet.OpCode, packet.Cid)
		u.ReadQueue <- packet
	}
	c.Close()
}

func (u *Server) writeAddr(c *net.UDPConn, a *net.UDPAddr, send chan *gamedata.Packet) {
	for {
		//fmt.Printf("[UDP] WAITING FOR PACKET FROM: %v\n", a.String())
		select {
		case packet := <-send:
			//fmt.Printf("[UDP] Writing data for %v\n", a.String())
			out, err := proto.Marshal(packet)
			if err != nil {
				fmt.Printf("[UDP] Failed marshalling %v. Reason: %v\n", packet.OpCode, err.Error())
				continue
			}
			c.WriteToUDP(out, a)
			//fmt.Printf("[UDP] Wrote %v bytes to %v\n", len(out), a.String())
		}
	}
}

// Stop TODO
func (u *Server) Stop() error {
	err := u.conn.Close()
	if err != nil {
		return err
	}
	return nil
}
