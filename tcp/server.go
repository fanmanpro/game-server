package tcp

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
	ip        string
	port      string
	listener  *net.TCPListener
	ReadQueue chan *gamedata.Packet
	// can potentially remove having to make an accept channel because trust client should be embedded and there no reason to do it in the game-server
	AcceptQueue chan *client.AcceptTCPPacket
}

// NewServer initializes a new UDP server without starting it
func NewServer(ip string, port string) *Server {
	return &Server{ip: ip, port: port, ReadQueue: make(chan *gamedata.Packet, 10), AcceptQueue: make(chan *client.AcceptTCPPacket, 2)}
}

// Start TODO
func (t *Server) Start() error {
	addr, err := net.ResolveTCPAddr("tcp4", t.ip+":"+t.port)

	if err != nil {
		return err
	}

	// setup listener for incoming TCP connection
	t.listener, err = net.ListenTCP("tcp4", addr)
	if err != nil {
		return err
	}

	go t.accept()

	return nil
}

func (t *Server) accept() {
	for {
		c, err := t.listener.AcceptTCP()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go t.procConn(c)
	}
}

func (t *Server) procConn(c *net.TCPConn) {
	buffer := make([]byte, 1024)

	// like doing this in a small scope because each client gets one and its pointer is passed into the TCPClient upon acceptance
	send := make(chan *gamedata.Packet, 1)

	// we want to be able to tell the write loop when the connection will be close so it can stop writing to it
	stop := make(chan bool, 2)

	// acceptance loop
	for {
		fmt.Printf("[TCP] Awaiting to read for %v\n", c.RemoteAddr().String())
		len, err := c.Read(buffer)
		if err != nil {
			fmt.Printf("[TCP] Failed reading %v. Reason: %v\n", c.RemoteAddr().String(), err.Error())
			stop <- true
			break
		}

		fmt.Printf("[TCP] Read %v bytes for %v\n", len, c.RemoteAddr().String())

		packet := &gamedata.Packet{}
		err = proto.Unmarshal(buffer[:len], packet)
		if err != nil {
			log.Println(err)
		}

		if packet.OpCode == gamedata.Header_ClientAppeal {
			acceptPacket := &client.AcceptTCPPacket{
				CID:  packet.Cid,
				Conn: c,
				Send: send,
			}
			t.AcceptQueue <- acceptPacket
		}
		break
	}

	//write loop
	go t.writeConn(c, send, stop)

	// read loop
	for {
		select {
		// this is hard to following but actually really cool to do :)
		case first := <-stop:
			if first {
				stop <- false
				return
			} else {
				c.Close()
				return
			}
		default:
			{
				len, err := c.Read(buffer)
				fmt.Printf("[TCP] Read %v bytes for %v\n", len, c.RemoteAddr().String())

				if err != nil {
					fmt.Printf("[TCP] Failed reading %v. Reason: %v\n", c.RemoteAddr().String(), err.Error())
					return
				}

				// or check if it is a close message type

				packet := &gamedata.Packet{}
				err = proto.Unmarshal(buffer[:len], packet)
				if err != nil {
					log.Println(err)
				}

				t.ReadQueue <- packet
			}
			break
		}
	}
}

func (t *Server) writeConn(c *net.TCPConn, send chan *gamedata.Packet, stop chan bool) {
	for {
		select {
		case first := <-stop:
			if first {
				stop <- false
				return
			} else {
				c.Close()
				return
			}
		case packet := <-send:
			//fmt.Printf("[TCP] Writing data for %v\n", c.RemoteAddr().String())
			out, err := proto.Marshal(packet)
			if err != nil {
				fmt.Printf("[TCP] Failed marshalling %v. Reason: %v\n", packet.OpCode, err.Error())
				continue
			}
			c.Write(out)
			fmt.Printf("[TCP] Wrote %v bytes for %v\n", len(out), c.RemoteAddr().String())
			break
		}
	}
}

// Stop TODO
func (t *Server) Stop() error {
	err := t.listener.Close()
	if err != nil {
		return err
	}
	return nil
}
