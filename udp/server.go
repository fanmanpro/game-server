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

	// can move this later when reconnecting is possible
	//acceptedAddr := make([]*net.UDPAddr, 1)

	// like doing this in a small scope because each client gets one and its pointer is passed into the TCPClient upon acceptance

	//these channel don't make any sense. the problem is that there is one send channel, but procAddr does
	//multiple "listens" to the same channel, so it takes a random one and not the one related to the address
	//so make the "acceptedAddr" and struct with the send and stop channel to make it manageble. also think
	//of how the start and started variables should work
	//REMINDERS
	//there are TWO procConn goroutines active, each should have they own send, stop, and start (maybe not, think!)
	//there are 2+ procAddr goroutines active, consider the above, take it from there

	for {
		unknown := true
		//fmt.Printf("[UDP] Awaiting to read\n")
		l, addr, err := c.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("[UDP] Failed reading %v. Reason: %v\n", addr.String(), err.Error())
			break
		}

		for _, a := range u.acceptedAddrs {
			//fmt.Printf("[UDP] ---------------------------- Accept Addr: %v vs %v\n", a, addr)
			if a.String() == addr.String() {
				unknown = false
				break
			}
		}
		if unknown {
			u.AcceptAddress(addr)
			//fmt.Printf("[UDP] ---------------------------- Accepting Addr: %v\n", addr)
			//go u.writeAddr(c, addr, send)
			//continue
		}

		//fmt.Printf("[UDP] Reading %v bytes for %v\n", l, addr.String())
		packet := &gamedata.Packet{}
		err = proto.Unmarshal(buffer[:l], packet)
		if err != nil {
			log.Println(err)
		}

		if unknown {
			unknown = false
			send := make(chan *gamedata.Packet, 10)
			go u.writeAddr(c, addr, send)
			acceptPacket := &client.AcceptUDPPacket{
				CID:  packet.Cid,
				Addr: addr,
				Send: send,
			}
			//fmt.Printf("[UDP] Reading was accept\n")
			u.AcceptQueue <- acceptPacket
			//fmt.Printf("[UDP] Read accept\n")
		}

		u.ReadQueue <- packet
		//fmt.Printf("[UDP] Read %v bytes for %v\n", l, addr.String())
		//if packet.OpCode == gamedata.Header_ClientDatagramAddress {
		//	found := false
		//	for _, a := range acceptedAddr {
		//		if a.String() == addr.String() {
		//			found = true
		//		}
		//	}
		//	if found {
		//		continue
		//	}
		//	acceptedAddr = append(acceptedAddr, addr)
		//	//fmt.Println("Not sim", addr)
		//	acceptPacket := &client.AcceptUDPPacket{
		//		CID:   packet.Cid,
		//		Addr:  addr,
		//		Send:  send,
		//		Start: start,
		//	}
		//	u.AcceptQueue <- acceptPacket
		//	// check if it was everyone
		//}
	}
	fmt.Println("CLLLLLLLLLLLLLLLLLOSED!")
	c.Close()
}

func (u *Server) writeAddr(c *net.UDPConn, a *net.UDPAddr, send chan *gamedata.Packet) {
	for {
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

// happens per addr (including the simulation)
//func (u *Server) procAddr(c *net.UDPConn, a *net.UDPAddr, send chan *gamedata.Packet, stop chan bool) {
//	buffer := make([]byte, 1024)
//
//	go u.writeAddr(c, a, send, stop)
//
//	// read loop
//	for {
//		select {
//		// this is hard to following but actually really cool to do :)
//		case first := <-stop:
//			if first {
//				stop <- false
//				return
//			} else {
//				// remove addr from accept addr list
//				return
//			}
//		default:
//			{
//				fmt.Printf("[UDP] MAIN Awaiting to read\n")
//				l, addr, err := c.ReadFromUDP(buffer)
//				//if addr.String() != a.String() {
//				//	fmt.Printf("[UDP] Read invalid %v bytes for %v\n", l, addr.String())
//				//	continue
//				//}
//				fmt.Printf("[UDP] MAIN Read %v bytes for %v\n", l, addr.String())
//
//				if err != nil {
//					fmt.Printf("[UDP] MAIN Failed reading %v. Reason: %v\n", addr.String(), err.Error())
//					return
//				}
//
//				// or check if it is a close message type
//
//				packet := &gamedata.Packet{}
//				err = proto.Unmarshal(buffer[:l], packet)
//				if err != nil {
//					log.Println(err)
//				}
//
//				//if u.simConn == c {
//				//u.ReadQueueSim <- packet
//				//} else {
//				u.ReadQueue <- packet
//				//}
//			}
//			break
//		}
//	}
//}

// Stop TODO
func (u *Server) Stop() error {
	err := u.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

//		case gamedata.Header_Velocity:
//			{
//				u.gameServer.Broadcast(packet)
//				return nil
//			}
//		case gamedata.Header_SimulationState:
//			{
//				//get it out to all clients
//				u.gameServer.Broadcast(packet)
//				return nil
//			}
//		//case gamedata.Header_Rigidbody:
//		//	{
//		//		out, err := proto.Marshal(packet)
//		//		if err != nil {
//		//			return err
//		//		}
//		//		//get it out to all clients
//		//		u.gameServer.Broadcast(&out)
//		//		return nil
//		//	}
//		case gamedata.Header_Force:
//			{
//				//get it out to all clients
//				u.gameServer.Broadcast(packet)
//				return nil
//			}
//		}
//	} else {
//		// from client
//		switch packet.Header.OpCode {
//		case gamedata.Header_ClientJoin: // TODO: Turn into TCP packet
//			{
//				phase := u.gameServer.Phase()
//				if phase != gs.Connect {
//					return errors.New("Game server not receiving client connections yet")
//				}
//
//				err := u.gameServer.SetClientAddress(packet.Header.Cid, addr)
//				if err != nil {
//					return err
//				}
//				u.gameServer.SendToSimulationServer(packet)
//				return nil
//			}
//		case gamedata.Header_Velocity:
//			{
//				u.gameServer.SendToSimulationServer(packet)
//				return nil
//			}
//		}
//	}
//	return errors.New("Invalid packet received from UDP server")
//}
