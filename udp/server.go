package udp

import (
	"fmt"
	"log"
	"net"

	"github.com/fanmanpro/coordinator-server/client"
	"github.com/fanmanpro/coordinator-server/gamedata"
	"github.com/golang/protobuf/proto"
)

var packetQueue []gamedata.Packet

// UDPPacket same as gamedata.Packet but with the addr include the validate in game-server.go
type UDPPacket struct {
	UDPAddr *net.UDPAddr
	Packet  *gamedata.Packet
}

// Server to open web socket ports
type Server struct {
	ip      string
	port    string
	simPort string
	conn    *net.UDPConn
	simConn *net.UDPConn

	simAcceptedAddr *net.UDPAddr

	ReadQueue    chan *gamedata.Packet
	ReadQueueSim chan *gamedata.Packet

	AcceptQueue chan *client.AcceptUDPPacket
}

// NewServer initializes a new UDP server without starting it
func NewServer(ip string, port string, simPort string) *Server {
	return &Server{ip: ip, port: port, simPort: simPort, ReadQueue: make(chan *gamedata.Packet, 1), ReadQueueSim: make(chan *gamedata.Packet, 1), AcceptQueue: make(chan *client.AcceptUDPPacket, 1)}
}

// Start starts the already intialized UDPServer
func (u *Server) Start() error {
	addr, err := net.ResolveUDPAddr("udp4", u.ip+":"+u.port)
	if err != nil {
		return err
	}

	simAddr, err := net.ResolveUDPAddr("udp4", u.ip+":"+u.simPort)
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

	u.simConn, err = net.ListenUDP("udp", simAddr)
	if err != nil {
		return err
	}
	//go u.readConn(u.simConn, stop, u.queuePacketSim)

	start := make(chan bool, 2)
	go u.procConn(u.conn, start)
	go u.procConn(u.simConn, start)

	//go u.writeConn(stop)

	return nil
}

func (u *Server) writeAddr(c *net.UDPConn, a *net.UDPAddr, send chan *gamedata.Packet, stop chan bool) {
	for {
		if u.simConn == c {
			fmt.Printf("[UDP] SIM Awaiting write for %v\n", a.String())
		} else {
			fmt.Printf("[UDP] CLIENT Awaiting write for %v\n", a.String())
		}
		select {
		case first := <-stop:
			if first {
				stop <- false
				return
			} else {
				return
			}
		case packet := <-send:
			fmt.Printf("[UDP] Writing data for %v\n", a.String())
			out, err := proto.Marshal(packet)
			if err != nil {
				fmt.Printf("[UDP] Failed marshalling %v. Reason: %v\n", packet.OpCode, err.Error())
				continue
			}
			c.WriteToUDP(out, a)
			fmt.Printf("[UDP] Wrote %v bytes for %v\n", len(out), a.String())
			break
		}
	}
}

// happens per addr (including the simulation)
func (u *Server) procAddr(c *net.UDPConn, a *net.UDPAddr, send chan *gamedata.Packet, stop chan bool) {
	buffer := make([]byte, 1024)

	go u.writeAddr(c, a, send, stop)

	// read loop
	for {
		select {
		// this is hard to following but actually really cool to do :)
		case first := <-stop:
			if first {
				stop <- false
				return
			} else {
				// remove addr from accept addr list
				return
			}
		default:
			{
				fmt.Printf("[UDP] MAIN Awaiting to read\n")
				l, addr, err := c.ReadFromUDP(buffer)
				//if addr.String() != a.String() {
				//	fmt.Printf("[UDP] Read invalid %v bytes for %v\n", l, addr.String())
				//	continue
				//}
				fmt.Printf("[UDP] MAIN Read %v bytes for %v\n", l, addr.String())

				if err != nil {
					fmt.Printf("[UDP] MAIN Failed reading %v. Reason: %v\n", addr.String(), err.Error())
					return
				}

				// or check if it is a close message type

				packet := &gamedata.Packet{}
				err = proto.Unmarshal(buffer[:l], packet)
				if err != nil {
					log.Println(err)
				}

				if u.simConn == c {
					u.ReadQueueSim <- packet
				} else {
					u.ReadQueue <- packet
				}
			}
			break
		}
	}
}

// happens per open UDP socket
func (u *Server) procConn(c *net.UDPConn, start chan bool) {
	buffer := make([]byte, 1024)

	// can move this later when reconnecting is possible
	acceptedAddr := make([]*net.UDPAddr, 1)

	// like doing this in a small scope because each client gets one and its pointer is passed into the TCPClient upon acceptance
	send := make(chan *gamedata.Packet, 1)

	//these channel don't make any sense. the problem is that there is one send channel, but procAddr does
	//multiple "listens" to the same channel, so it takes a random one and not the one related to the address
	//so make the "acceptedAddr" and struct with the send and stop channel to make it manageble. also think
	//of how the start and started variables should work
	//REMINDERS
	//there are TWO procConn goroutines active, each should have they own send, stop, and start (maybe not, think!)
	//there are 2+ procAddr goroutines active, consider the above, take it from there

	// we want to be able to tell the write loop when the connection will be close so it can stop writing to it
	stop := make(chan bool, 1)

	var started bool
	started = false

	// acceptance loop
	for {
		if started {
			break
		}
		select {
		case <-start:
			{
				if u.simConn == c {
					go u.procAddr(c, u.simAcceptedAddr, send, stop)
				} else {
					for _, a := range acceptedAddr {
						if a != nil && a != u.simAcceptedAddr {
							go u.procAddr(c, a, send, stop)
						}
					}
				}
				started = true
				continue
			}
		default:
			{
				fmt.Printf("[UDP] ACCEPT Awaiting to read\n")
				l, addr, err := c.ReadFromUDP(buffer)
				if err != nil {
					fmt.Printf("[UDP] ACCEPT Failed reading %v. Reason: %v\n", addr.String(), err.Error())
					//stop <- true
					break
				}

				if u.simConn == c {
					fmt.Printf("[UDP] ACCEPT SIM Read %v bytes for %v\n", l, addr.String())
				} else {
					fmt.Printf("[UDP] ACCEPT CLIENT Read %v bytes for %v\n", l, addr.String())
				}

				packet := &gamedata.Packet{}
				err = proto.Unmarshal(buffer[:l], packet)
				if err != nil {
					log.Println(err)
				}

				if packet.OpCode == gamedata.Header_ClientDatagramAddress {
					if u.simConn == c {
						if u.simAcceptedAddr != nil {
							continue
						}
						u.simAcceptedAddr = addr
						//fmt.Println("Sim", addr)
						acceptPacket := &client.AcceptUDPPacket{
							CID:   packet.Cid,
							Addr:  addr,
							Send:  send,
							Start: start,
						}
						u.AcceptQueue <- acceptPacket
					} else {
						found := false
						for _, a := range acceptedAddr {
							if a.String() == addr.String() {
								found = true
							}
						}
						if found {
							continue
						}
						acceptedAddr = append(acceptedAddr, addr)
						//fmt.Println("Not sim", addr)
						acceptPacket := &client.AcceptUDPPacket{
							CID:   packet.Cid,
							Addr:  addr,
							Send:  send,
							Start: start,
						}
						u.AcceptQueue <- acceptPacket
					}
					// check if it was everyone
				}
			}
		}
	}
	<-stop
	c.Close()
}

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
