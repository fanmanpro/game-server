package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

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
	return &Server{}
	// return &Server{rate: 50, tick: 1, stop: make(chan bool)} // 100 is 10 ticks per second, 50 is 20, 33 is 30, etc.
}

var simulationTCPConnection *net.TCPConn
var simulationTCPAddressHost *net.TCPAddr
var simulationTCPAddressRemote *net.TCPAddr

var clientTCPConnection *net.TCPConn
var clientTCPAddressHost *net.TCPAddr
var clientTCPAddressRemote *net.TCPAddr

var simulationUDPConnection *net.UDPConn
var simulationUDPAddressHost *net.UDPAddr
var simulationUDPAddressRemote *net.UDPAddr

var clientUDPConnection *net.UDPConn
var clientUDPAddressHost *net.UDPAddr
var clientUDPAddressRemote *net.UDPAddr

var seats map[string]*net.TCPAddr

const rate time.Duration = 50

var tick int32 = 1

var sent time.Time
var end time.Time

func connectSimulationTCP() {
	var err error

	simulationTCPAddressHost, err = net.ResolveTCPAddr(tcpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "9999"))
	if err != nil {
		fmt.Println(err)
	}

	simulationTCPAddressRemote, err = net.ResolveTCPAddr(tcpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "1999"))
	if err != nil {
		fmt.Println(err)
	}

	// connect the sim
	for {
		simulationTCPConnection, err = net.DialTCP(tcpNetwork, simulationTCPAddressHost, simulationTCPAddressRemote)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if err == nil {
			fmt.Println("simulation connected")
			break
		}
		time.Sleep(time.Second)
	}
}

func connectClientsTCP() {
	var err error

	clientTCPAddressHost, err = net.ResolveTCPAddr(tcpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "9889"))
	if err != nil {
		fmt.Println(err)
	}

	// this should happen per client
	func() {
		clientTCPAddressRemote, err = net.ResolveTCPAddr(tcpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "1889"))
		if err != nil {
			fmt.Println(err)
		}

		// once the game server knows about the client, it has 5 attempts/seconds to successfully connect
		attempts := 5
		for {
			clientTCPConnection, err = net.DialTCP(tcpNetwork, clientTCPAddressHost, clientTCPAddressRemote)
			attempts--
			if err == nil {
				fmt.Println("client connected")
				break
			}
			if attempts < 1 {
				fmt.Println("client failed to connect")
				break
			}
			time.Sleep(time.Second)
		}
	}()
}

func configureSeats() {
	buffer := make([]byte, 64*1024)
	l, err := simulationTCPConnection.Read(buffer)
	fmt.Println("configuring seats")
	if err != nil {
		fmt.Println(err)
	}

	incoming := &serializable.Packet{}
	err = proto.Unmarshal(buffer[0:l], incoming)
	if err != nil {
		fmt.Println(err)
	}

	if incoming.OpCode == serializable.Packet_SeatConfiguration {
		seatConfiguration := &serializable.SeatConfiguration{}
		err := proto.Unmarshal(incoming.Data.GetValue(), seatConfiguration)
		fmt.Printf("seating %+v\n", seatConfiguration)
		if err != nil {
			fmt.Printf("Invalid seat configuration packet: %v\n", err)
		}
		seatCount := len(seatConfiguration.Seats)
		fmt.Printf("Server registered %v seats from simulation\n", seatCount)
		// seatsByTCP = make(map[*net.UDPAddr]*net.TCPAddr, seatCount)
		seats = make(map[string]*net.TCPAddr, seatCount)
		for i, seat := range seatConfiguration.Seats {
			// fmt.Printf("seating at %v for %v", nil, clientTCPAddressRemote.Port)
			if i == 1 {
				continue
			}
			seats[seat.GUID] = clientTCPAddressRemote
		}
	}
}

func connectClientsUDP() {
	clientUDPAddressHost, err := net.ResolveUDPAddr(udpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "9888")) // this should be a client address, not localhost
	if err != nil {
		fmt.Println(err)
	}

	clientUDPAddressRemote, err := net.ResolveUDPAddr(udpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "1888")) // this should be a client address, not localhost
	if err != nil {
		fmt.Println(err)
	}

	clientUDPConnection, err = net.DialUDP(udpNetwork, clientUDPAddressHost, clientUDPAddressRemote)
	if err != nil {
		fmt.Println(err)
	}
}

func connectSimulationUDP() {
	simulationUDPAddressHost, err := net.ResolveUDPAddr(udpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "9998"))
	if err != nil {
		fmt.Println(err)
	}

	simulationUDPAddressRemote, err := net.ResolveUDPAddr(udpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "1998"))
	if err != nil {
		fmt.Println(err)
	}

	simulationUDPConnection, err = net.DialUDP(udpNetwork, simulationUDPAddressHost, simulationUDPAddressRemote)
	if err != nil {
		fmt.Println(err)
	}
}

func updateSimulation() {
	for {
		end = time.Now().Add(rate * time.Millisecond)

		// set the read deadline for when we are waiting for the returned tick
		err := simulationUDPConnection.SetReadDeadline(end)
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}

		// wait for a context to come back from the sim
		buffer := make([]byte, 64*1024*1024)
		l, _, err := simulationUDPConnection.ReadFromUDP(buffer)
		if err != nil {
			//if n != 0 || err == nil || !err.(Error).Timeout() {
			// fmt.Printf("%v\n", err)
			if !err.(net.Error).Timeout() {
				fmt.Printf("%+v\n", err)
			}
			continue
		}

		// unmarshal the incoming context
		incoming := &serializable.Context3D{}
		err = proto.Unmarshal(buffer[0:l], incoming)
		if err != nil {
			fmt.Printf("%v\n", err)
		}

		// check if the incoming tick is the same as the current expected tick
		if incoming.Tick != tick {
			fmt.Printf("context was dropped: %v vs %v (old)\n", incoming.Tick, tick)
			//return
		} else {
			clientUDPConnection.Write(buffer[0:l])
		}

		// define the outgoing packet we want to send to the sim
		tick++

		outgoing := &serializable.Context3D{
			Tick: tick,
		}

		// marshal the outgoing packet
		data, err := proto.Marshal(outgoing)
		if err != nil {
			fmt.Printf("%v\n", err)
			continue
		}

		time.Sleep(end.Sub(time.Now()))

		// write the outgoing packet to the socket shared with the sim
		simulationUDPConnection.Write(data)
	}
}

func broadcastSeats() {
	for g, a := range seats {
		if a == nil {
			fmt.Println("address was nil")
			continue
		}
		fmt.Printf("%v seated at %v\n", a.Port, g)

		ss := &serializable.Seat{
			GUID: g,
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

		_, err = clientTCPConnection.Write(data)
		if err != nil {
			if err.(net.Error).Timeout() {
				fmt.Printf("client disconnected\n")
			}
		}
		break
	}
}

func runSimulation() {
	fmt.Println("sending run simulation packet")
	runSimulationPacket := &serializable.Packet{OpCode: serializable.Packet_RunSimulation}
	data, err := proto.Marshal(runSimulationPacket)
	if err != nil {
		fmt.Println(err)
	}
	simulationTCPConnection.Write(data)
}

func receiveClientsUDP() {
	if clientUDPConnection == nil {
		fmt.Println("no UDP connection for client to receive with")
		return
	} else {
		fmt.Println("receiving UDP for client")
	}
	for {
		// wait for context
		buffer := make([]byte, 64*1024*1024)
		l, _, err := clientUDPConnection.ReadFromUDP(buffer)
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

		simulationUDPConnection.Write(outgoing)
	}
}
func receiveClientsTCP() {
	if clientTCPConnection == nil {
		fmt.Println("no TCP connection for client to receive with")
		return
	} else {
		fmt.Println("receiving TCP for client")
	}
	defer clientTCPConnection.Close()
	for {
		// wait for context
		buffer := make([]byte, 64*1024*1024)
		_, err := clientTCPConnection.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("client disconnected")
			} else {
				if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
					fmt.Printf("client timed out")
				} else {
					fmt.Printf("client socket error: %v", err)
				}
			}
			break
		}
		// unmarshal context from client. validate then forward.
		// data := buffer[:l]
		// packet := &serializable.Packet{}
		// err = proto.Unmarshal(data, packet)
		// if err != nil {
		// 	log.Println(err)
		// }
	}
}

func receiveSimulationTCP() {
	fmt.Printf("receiving TCP for simulation\n")
	defer simulationTCPConnection.Close()
	for {
		// wait for context
		buffer := make([]byte, 64*1024*1024)
		_, err := simulationTCPConnection.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("simulation disconnected")
			} else {
				if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
					fmt.Printf("simulation timed out")
				} else {
					fmt.Printf("simulation socket error: %v", err)
				}
			}
			break
		}
		// unmarshal context from client. validate then forward.
		// data := buffer[:l]
		// packet := &serializable.Packet{}
		// err = proto.Unmarshal(data, packet)
		// if err != nil {
		// 	log.Println(err)
		// }
	}
}

func (s *Server) Start() error {
	fmt.Println("starting game server")

	restart := make(chan int)

	think about what will happen if the simulation restarts. e.g. when the match ends.
	the simulation must be able to restart
	the client must be able to restart

	connectSimulationTCP()
	connectClientsTCP()

	configureSeats()
	broadcastSeats()

	go receiveSimulationTCP()
	go receiveClientsTCP()

	connectSimulationUDP()
	connectClientsUDP()

	go receiveClientsUDP()
	go updateSimulation()

	runSimulation()

	<-restart

	fmt.Println("Server ended")
	return nil
}
