package server

import (
	"fmt"
	"log"
	"net"
	"time"

	//"github.com/fanmanpro/coordinator-server/gamedata"

	"github.com/fanmanpro/game-server/serializable"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

const localhostAddress string = "127.0.0.1"
const tcpNetwork string = "tcp4"
const udpNetwork string = "udp4"

type Server struct {
	// TCP
	tcpSimSocket           *net.TCPConn
	tcpClientSocket        *net.TCPConn
	tcpLocalSimAddress     *net.TCPAddr
	tcpRemoteSimAddress    *net.TCPAddr
	tcpLocalClientAddress  *net.TCPAddr
	tcpRemoteClientAddress *net.TCPAddr

	// UDP
	udpSimSocket           *net.UDPConn
	udpClientSocket        *net.UDPConn
	udpLocalSimAddress     *net.UDPAddr
	udpRemoteSimAddress    *net.UDPAddr
	udpLocalClientAddress  *net.UDPAddr
	udpRemoteClientAddress *net.UDPAddr

	// Game Server
	rate       time.Duration
	tick       int32
	seater     Seater
	stop       chan bool
	seatsByTCP map[*net.UDPAddr]*net.TCPAddr
	seats      map[string]*net.TCPAddr
}

func NewServer() *Server {
	return &Server{rate: 50} // 100 is 10 ticks per second, 50 is 20, 33 is 30, etc.
}

func (s *Server) Start() error {
	fmt.Println("starting game server")
	s.tick = 1
	s.stop = make(chan bool)

	// ✓ open tcp sockets for sim and wait for it to connect
	// ✓ open tcp sockets for clients and wait for them to connect
	// ✓ listen for configuration packet from sim
	// ✓ create seats based on configuration
	// ✓ assign tcp addr to each client, populating the seats created from configuration

	// open udp sockets for sim
	// open udp sockets for clients

	// send tcp seat configuration packet to all clients (not sim)

	// send tcp packet to sim to unpause simuluation and open the flood gates

	err := s.openTCPSockets()
	if err != nil {
		return err
	}

	func() {
		fmt.Println("listening for TCP")
		for {
			// Make a buffer to hold incoming data.
			buffer := make([]byte, 64*1024)
			// Read the incoming connection into the buffer.
			l, err := s.tcpSimSocket.Read(buffer)
			if err != nil {
				if !err.(net.Error).Timeout() {
					fmt.Printf("%+v\n", err)
				}
				break
			}

			// unmarshal the incoming seat configuration
			incoming := &serializable.Packet{}
			readBuffer := buffer[0:l]
			err = proto.Unmarshal(readBuffer, incoming)
			if err != nil {
				fmt.Printf("%v\n", err)
				break
			}

			if incoming.OpCode == serializable.Packet_SeatConfiguration {
				seatConfiguration := &serializable.SeatConfiguration{}
				err = proto.Unmarshal(incoming.Data.Value, seatConfiguration)
				if err != nil {
					fmt.Printf("Invalid seat configuration packet: %v\n", err)
					break
				}
				seatCount := len(seatConfiguration.Seats)
				fmt.Printf("Server registered %v seats from simulation\n", seatCount)
				s.seatsByTCP = make(map[*net.UDPAddr]*net.TCPAddr, seatCount)
				s.seats = make(map[string]*net.TCPAddr, seatCount)

				// open the UDP sockets here so they are open once the clients receive they seats
				// REFACTOR: Start Function
				err = s.openUDPSockets()
				if err != nil {
					return
				}
				// start the server ticker. careful for this one.
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
				// REFACTOR: End Function

				for i, seat := range seatConfiguration.Seats {
					if i == 0 {
						continue
					}
					s.seats[seat.GUID] = s.tcpRemoteClientAddress
					fmt.Printf("%v seated at %v\n", s.tcpRemoteClientAddress.Port, seat.GUID)
					ss := &serializable.Seat{
						GUID: seat.GUID,
					}
					any, err := ptypes.MarshalAny(ss)
					if err != nil {
						fmt.Printf("%v\n", err)
						return
					}

					seatPacket :=
						&serializable.Packet{
							OpCode: serializable.Packet_Seat,
							Data:   any,
						}

					data, err := proto.Marshal(seatPacket)
					if err != nil {
						fmt.Printf("%v\n", err)
						return
					}
					s.tcpClientSocket.Write(data)
					break
				}
				return
			}
		}
	}()

	if err != nil {
		return err
	}

	runSimulationPacket := &serializable.Packet{OpCode: serializable.Packet_RunSimulation}
	data, err := proto.Marshal(runSimulationPacket)
	if err != nil {
		fmt.Printf("%v\n", err)
		return err
	}
	s.tcpSimSocket.Write(data)

	// send an initial packet so the simulation server can start its read/write loop
	// REFACTOR: Start Function
	outgoing := &serializable.Context3D{
		Tick: s.tick,
	}
	data, err = proto.Marshal(outgoing)
	if err != nil {
		fmt.Printf("%v\n", err)
		return nil
	}
	time.Sleep(end.Sub(time.Now()))
	s.udpSimSocket.Write(data)
	//s.udpSimSocket.WriteTo(data, s.udpSimSocket.RemoteAddr())
	// REFACTOR: End Function

	s.receiveClient()
	return nil
}
func (s *Server) Stop(err error) {
	fmt.Println("stopping game server")
	// s.stop <- true
	// s.tcpClientSocket.Close()
	// s.udpClientSocket.Close()
	// fmt.Printf("game server stopped (%v)\n", err)
}

func (s *Server) broadcast(data []byte) {
	//err := s.socketClient.SetWriteDeadline(time.Now().Add((s.rate) * time.Millisecond))
	//if err != nil {
	//	fmt.Printf("%v\n", err)
	//	return
	//}
	s.udpClientSocket.Write(data)
	// s.udpClientSocket.WriteTo(data, s.udpClientSocket.RemoteAddr())
}

var sent time.Time
var end time.Time

func (s *Server) update() {
	end = time.Now().Add(s.rate * time.Millisecond)

	// set the read deadline for when we are waiting for the returned tick
	err := s.udpSimSocket.SetReadDeadline(end)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	// wait for a context to come back from the sim
	buffer := make([]byte, 64*1024*1024)
	l, _, err := s.udpSimSocket.ReadFromUDP(buffer)
	if err != nil {
		//err = &OpError{Op: "read", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
		//if n != 0 || err == nil || !err.(Error).Timeout() {
		if !err.(net.Error).Timeout() {
			fmt.Printf("%+v\n", err)
		}
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
		s.broadcast(readBuffer)
	}

	//fmt.Printf("tick: %v (%vms trip) (%vms sleep)\n", s.tick, time.Now().Sub(start).Milliseconds(), end.Sub(time.Now()).Milliseconds())

	// define the outgoing packet we want to send to the sim
	s.tick++

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
	s.udpSimSocket.Write(data)
	// s.udpSimSocket.WriteToUDP(data, s.udpRemoteSimAddress)
	// s.udpSimSocket.WriteTo(data, s.udpSimSocket.RemoteAddr())
}

// received a context from either
// NO! GET RID OF THIS. FROM THE SIM WE READ AFTER WE WROTE AS A LOOP. NOT THIS SEPARATE READ LOOP.
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
		buffer := make([]byte, 64*1024*1024)
		l, _, err := s.udpClientSocket.ReadFromUDP(buffer)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		// unmarshal context from client. validate then forward.
		data := buffer[:l]
		context := &serializable.Context3D{}
		err = proto.Unmarshal(data, context)
		if err != nil {
			log.Println(err)
		}

		context.Client = true

		outgoing, err := proto.Marshal(context)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		// compare
		// fmt.Println("here")
		// fmt.Printf("%v | %v | v | v | v\n", s.udpClientSocket.RemoteAddr().String(), s.udpClientSocket.LocalAddr().String())
		s.udpSimSocket.Write(outgoing)
		// s.udpSimSocket.WriteTo(data, s.udpSimSocket.RemoteAddr())
	}
}

func (s *Server) openTCPSockets() error {
	var err error

	s.tcpLocalSimAddress, err = net.ResolveTCPAddr(tcpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "9999"))
	if err != nil {
		return err
	}

	s.tcpRemoteSimAddress, err = net.ResolveTCPAddr(tcpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "1999"))
	if err != nil {
		return err
	}

	s.tcpLocalClientAddress, err = net.ResolveTCPAddr(tcpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "9889")) // this should be a client address, not localhost
	if err != nil {
		return err
	}

	s.tcpRemoteClientAddress, err = net.ResolveTCPAddr(tcpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "1889")) // this should be a client address, not localhost
	if err != nil {
		return err
	}

	// connect the sim
	err = func() error {
		attempts := 2
		for {
			s.tcpSimSocket, err = net.DialTCP(tcpNetwork, s.tcpLocalSimAddress, s.tcpRemoteSimAddress)
			attempts--
			if err == nil {
				fmt.Println("simulation connected")
				return nil
			}
			if attempts < 1 {
				break
			}
			time.Sleep(time.Second)
		}
		return err
	}()

	for {
		reloadSimulationPacket := &serializable.Packet{OpCode: serializable.Packet_ReloadSimulation}
		data, err := proto.Marshal(reloadSimulationPacket)
		if err != nil {
			fmt.Printf("%v\n", err)
			return err
		}
		s.tcpSimSocket.Write(data)
		time.Sleep(time.Second * 5)
	}

	if err != nil {
		return err
	}

	// connect the clients
	err = func() error {
		attempts := 2
		for {
			s.tcpClientSocket, err = net.DialTCP(tcpNetwork, s.tcpLocalClientAddress, s.tcpRemoteClientAddress)
			attempts--
			if err == nil {
				fmt.Println("client connected")
				return nil
			}
			if attempts < 1 {
				break
			}
			time.Sleep(time.Second)
		}
		return err
	}()

	if err != nil {
		return err
	}

	return nil
}
func (s *Server) openUDPSockets() error {
	var err error

	s.udpLocalSimAddress, err = net.ResolveUDPAddr(udpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "9998"))
	if err != nil {
		return err
	}

	s.udpRemoteSimAddress, err = net.ResolveUDPAddr(udpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "1998"))
	if err != nil {
		return err
	}

	s.udpLocalClientAddress, err = net.ResolveUDPAddr(udpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "9888")) // this should be a client address, not localhost
	if err != nil {
		return err
	}

	s.udpRemoteClientAddress, err = net.ResolveUDPAddr(udpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "1888")) // this should be a client address, not localhost
	if err != nil {
		return err
	}

	s.udpSimSocket, err = net.DialUDP(udpNetwork, s.udpLocalSimAddress, s.udpRemoteSimAddress)
	if err != nil {
		return err
	}

	s.udpClientSocket, err = net.DialUDP(udpNetwork, s.udpLocalClientAddress, s.udpRemoteClientAddress)
	if err != nil {
		return err
	}

	return nil
}
