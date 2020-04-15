package server

import (
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"github.com/fanmanpro/game-server/serializable"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

const localhostAddress string = "127.0.0.1"
const tcpNetwork string = "tcp4"
const udpNetwork string = "udp4"

type Server struct {
}

func NewServer() *Server {
	return &Server{}
	// return &Server{rate: 50, tick: 1, stop: make(chan bool)} // 100 is 10 ticks per second, 50 is 20, 33 is 30, etc.
}

var simulationTCPConnection *net.TCPConn
var simulationTCPAddressHost *net.TCPAddr
var simulationTCPAddressRemote *net.TCPAddr

var clientTCPConnections []*net.TCPConn
var clientTCPAddressHost *net.TCPAddr

var simulationUDPConnection *net.UDPConn
var simulationUDPAddressHost *net.UDPAddr
var simulationUDPAddressRemote *net.UDPAddr

var clientUDPConnections []*net.UDPConn
var clientUDPAddressHost *net.UDPAddr

var seats map[string]*net.TCPConn

const rate time.Duration = 50

var tick int32 = 1

var sent time.Time
var end time.Time

func connectSimulationTCP() error {
	var err error

	simulationTCPAddressHost, err = net.ResolveTCPAddr(tcpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "9999"))
	if err != nil {
		return err
	}

	simulationTCPAddressRemote, err = net.ResolveTCPAddr(tcpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "1999"))
	if err != nil {
		return err
	}

	// connect the sim
	for {
		simulationTCPConnection, err = net.DialTCP(tcpNetwork, simulationTCPAddressHost, simulationTCPAddressRemote)
		if err != nil {
			fmt.Println("sim tcp", err)
		} else {
			return nil
		}
		time.Sleep(time.Second)
	}
}

func connectClientsTCPAsync() {
	var err error

	clientTCPAddressHost, err = net.ResolveTCPAddr(tcpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "9889"))
	if err != nil {
		fmt.Println(err)
		return
	}

	// clientTCPConnections = make([]*net.TCPConn, 0)

	// this should happen per client
	listener, err := net.ListenTCP("tcp", clientTCPAddressHost)
	if err != nil {
		fmt.Println(err)
		return
	}
	listener.SetDeadline(time.Now().Add(1 * time.Second))

	seatCount := len(seats)
	for i := 0; i < seatCount; i++ {
		clientTCPConnection, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println("only", i, "client(s) managed to connected: ", err)
			seatsConn <- nil
			return
		}
		fmt.Println("client connected")
		clientTCPConnections = append(clientTCPConnections, clientTCPConnection)
		seatsConn <- clientTCPConnection
		go receiveClientTCP(clientTCPConnection)
	}
	return
}

var seatsConn chan *net.TCPConn

func fillSeats() error {
	seatsConn = make(chan *net.TCPConn)
	for guid := range seats {
		tcpConnection := <-seatsConn
		if tcpConnection != nil {
			fmt.Printf("seating at %v for %v\n", guid, tcpConnection)
			seats[guid] = tcpConnection
		}
	}
	return nil
}
func configureSeats() error {
	var err error
	var l int

	buffer := make([]byte, 64*1024)
	l, err = simulationTCPConnection.Read(buffer)
	fmt.Println("configuring seats")
	if err != nil {
		return err
	}

	incoming := &serializable.Packet{}
	err = proto.Unmarshal(buffer[0:l], incoming)
	if err != nil {
		return err
	}

	if incoming.OpCode == serializable.Packet_SeatConfiguration {
		seatConfiguration := &serializable.SeatConfiguration{}
		err := proto.Unmarshal(incoming.Data.GetValue(), seatConfiguration)
		if err != nil {
			return err
		}
		fmt.Printf("server registered %v seats from simulation\n", len(seatConfiguration.Seats))
		seats = make(map[string]*net.TCPConn)
		for _, seat := range seatConfiguration.Seats {
			fmt.Printf("seating reserved for %v\n", seat.GUID)
			seats[seat.GUID] = nil
		}
	} else {
		return errors.New("invalid seat configuration packet opcode")
	}
	return nil
}

func connectClientsUDP() error {
	var err error

	clientUDPAddressHost, err = net.ResolveUDPAddr(udpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "9888")) // this should be a client address, not localhost
	if err != nil {
		return err
	}

	for _, tcpConnection := range seats {
		if tcpConnection == nil {
			continue
		}
		var clientUDPAddressRemote *net.UDPAddr
		clientUDPAddressRemote, err = net.ResolveUDPAddr(udpNetwork, fmt.Sprintf("%v:%v", strings.Split(tcpConnection.RemoteAddr().String(), ":")[0], "1888")) // this should be a client address, not localhost
		if err != nil {
			return err
		}

		clientUDPConnection, err := net.DialUDP(udpNetwork, clientUDPAddressHost, clientUDPAddressRemote)
		if err != nil {
			return err
		}

		clientUDPConnections = append(clientUDPConnections, clientUDPConnection)
		go receiveClientUDP(clientUDPConnection)
	}

	return nil
}

func connectSimulationUDP() error {
	simulationUDPAddressHost, err := net.ResolveUDPAddr(udpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "9998"))
	if err != nil {
		return err
	}

	simulationUDPAddressRemote, err := net.ResolveUDPAddr(udpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "1998"))
	if err != nil {
		return err
	}

	simulationUDPConnection, err = net.DialUDP(udpNetwork, simulationUDPAddressHost, simulationUDPAddressRemote)
	if err != nil {
		return err
	}

	simulationDisconnected = false

	fmt.Println("connected UDP for simulation")
	return nil
}

func receiveSimulationUDPAsync() {
	defer disconnectSimulation()
	fmt.Println("receiving UDP for simulation")
	for simulationUDPConnection != nil {
		end = time.Now().Add(rate * time.Millisecond)

		// set the read deadline for when we are waiting for the returned tick
		err := simulationUDPConnection.SetReadDeadline(end)
		if err != nil {
			fmt.Println(err)
			return
		}

		// wait for a context to come back from the sim
		buffer := make([]byte, 64*1024*1024)
		l, _, err := simulationUDPConnection.ReadFromUDP(buffer)
		if err != nil {
			if !err.(net.Error).Timeout() {
				fmt.Println(err)
				return
			}
			// its ok if its a timeout
			// fmt.Printf("sim read timout")
			continue
		}
		// fmt.Printf("sim read")
		// fmt.Println(tick)

		// unmarshal the incoming context
		incoming := &serializable.Context3D{}
		err = proto.Unmarshal(buffer[0:l], incoming)
		if err != nil {
			fmt.Printf("sim proto: %v\n", err)
		}

		// check if the incoming tick is the same as the current expected tick
		if incoming.Tick != tick {
			fmt.Printf("context was dropped: %v vs %v (old)\n", incoming.Tick, tick)
			continue
			//return
		} else {
			for _, c := range clientUDPConnections {
				_, err := c.Write(buffer[0:l])
				if err != nil {
					fmt.Printf("%v\n", err)
					continue
				}
			}
		}

		// define the outgoing packet we want to send to the sim
		tick++

		outgoing := &serializable.Context3D{
			Tick: tick,
		}

		// marshal the outgoing packet
		data, err := proto.Marshal(outgoing)
		if err != nil {
			fmt.Println(err)
			return
			// fmt.Printf("%v\n", err)
			// continue
		}

		time.Sleep(end.Sub(time.Now()))

		// write the outgoing packet to the socket shared with the sim

		// refactor this with channels and pausing this read loop
		if simulationUDPConnection != nil {
			_, err = simulationUDPConnection.Write(data)
			if err != nil {
				fmt.Println(err)
				return
				// fmt.Printf("%v\n", err)
				// break
			}
			// fmt.Println("wrote to sim")
		}
	}
	return
}

func broadcastSeats() error {
	for g, c := range seats {
		if c == nil {
			fmt.Println("address was nil")
			continue
		}
		// fmt.Printf("%v seated at %v\n", a.Port, g)

		ss := &serializable.Seat{
			GUID: g,
		}
		any, err := ptypes.MarshalAny(ss)
		if err != nil {
			return err
		}

		seatPacket :=
			&serializable.Packet{
				OpCode: serializable.Packet_Seat,
				Data:   any,
			}

		data, err := proto.Marshal(seatPacket)
		if err != nil {
			return err
		}

		_, err = c.Write(data)
		if err != nil {
			if err.(net.Error).Timeout() {
				fmt.Printf("client disconnected 2\n")
				return nil
			}
			return err
		}
		// break
	}
	return nil
}

func runSimulation() error {
	fmt.Println("sending run simulation packet")
	ss := &serializable.RunSimulation{
		Tick: tick,
	}
	any, err := ptypes.MarshalAny(ss)
	if err != nil {
		return err
		// fmt.Printf("%v\n", err)
		// return
	}
	runSimulationPacket := &serializable.Packet{OpCode: serializable.Packet_RunSimulation, Data: any}
	data, err := proto.Marshal(runSimulationPacket)
	if err != nil {
		return err
		// fmt.Println(err)
	}
	simulationTCPConnection.Write(data)
	return nil
}

func receiveClientUDP(clientUDPConnection *net.UDPConn) error {
	if clientUDPConnection == nil {
		return errors.New("no UDP connection for client to receive with")
	}
	fmt.Println("receiving UDP for client")
	for {
		// wait for context
		buffer := make([]byte, 64*1024*1024)
		l, _, err := clientUDPConnection.ReadFromUDP(buffer)
		if err != nil {
			return err
		}

		// unmarshal context from client. validate then forward.
		data := buffer[:l]
		context := &serializable.Context3D{}
		err = proto.Unmarshal(data, context)
		if err != nil {
			return err
		}

		context.Client = true

		outgoing, err := proto.Marshal(context)
		if err != nil {
			return err
		}

		// refactor this with channels and pausing this read loop
		if simulationUDPConnection != nil {
			simulationUDPConnection.Write(outgoing)
		} else {
			fmt.Println("simulation udp connection is nil")
		}
	}
	return nil
}
func receiveClientTCP(clientTCPConnection *net.TCPConn) error {
	if clientTCPConnection == nil {
		return errors.New("no TCP connection for client to receive with")
	} else {
		fmt.Println("receiving TCP for client")
	}
	defer disconnectClient()
	for {
		// wait for context
		buffer := make([]byte, 64*1024*1024)
		_, err := clientTCPConnection.Read(buffer)
		if err != nil {
			if err == io.EOF {
				return errors.New("client disconnected")
			}
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				return errors.New("client timed out")
			}
			return err
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

func receiveSimulationTCPAsync() {
	fmt.Println("receiving TCP for simulation")

	defer disconnectSimulation()
	for {
		// wait for context
		buffer := make([]byte, 64*1024*1024)
		_, err := simulationTCPConnection.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Println("simulation disconnected")
				return
				// reconnectSimulation()
			}
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				fmt.Println("simulation timed out")
				return
			}
			fmt.Println(err)
			return
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

func disconnectClient() {
	// clientTCPConnection.SetLinger(0)
	// clientTCPConnection.SetNoDelay(false)
	// clientTCPConnection.Close()

	// clientUDPConnection.Close()

	fmt.Println("client connections closed")

	reconnectClient()
}

var simulationDisconnected bool

func disconnectSimulation() error {
	if simulationDisconnected {
		return errors.New("simulation already disconnected")
	}

	simulationDisconnected = true
	if simulationTCPConnection != nil {
		simulationTCPConnection.SetLinger(0)
		simulationTCPConnection.SetKeepAlive(false)
		simulationTCPConnection.Close()
		simulationTCPConnection = nil
	}

	if simulationUDPConnection != nil {
		simulationUDPConnection.Close()
		simulationUDPConnection = nil
	}

	if simulationTCPConnection == nil && simulationUDPConnection == nil {
		fmt.Println("simulation connections closed")

		reconnectSimulation()
	}
	return nil
}

func reconnectSimulation() error {
	var err error
	fmt.Println("simulation reconnecting")

	err = connectSimulationTCP()
	if err != nil {
		return err
	}

	// is it a reconnect or a restart? for now, its always a restart
	err = configureSeats()
	if err != nil {
		return err
	}
	err = broadcastSeats()
	if err != nil {
		return err
	}

	go receiveSimulationTCPAsync()

	err = connectSimulationUDP()
	if err != nil {
		return err
	}
	go receiveSimulationUDPAsync()

	err = runSimulation()
	if err != nil {
		return err
	}

	return nil
}

func reconnectClient() {
	fmt.Println("client reconnecting")

	// connectClientsTCP()

	// is it a reconnect or a restart? for now, its always a restart
	// configureSeats()

	// connectSimulationUDP()
	// go receiveSimulationUDP()

	// runSimulation()
}

func (s *Server) Start() error {
	var err error
	fmt.Println("starting game server")

	restart := make(chan int)

	// connect the simulation
	err = connectSimulationTCP()
	if err != nil {
		return err
	}
	fmt.Println("simulation connected")

	// get how many clients can connect and reserve seats
	err = configureSeats()
	if err != nil {
		return err
	}
	fmt.Println("seats configured")

	// allow clients to connect based on seat configuration
	go connectClientsTCPAsync()

	// fill seats as client connections get accepted
	err = fillSeats()
	if err != nil {
		return err
	}
	fmt.Println("clients connected")

	// send each client the seats
	err = broadcastSeats()
	if err != nil {
		return err
	}

	// allow simulation to send tcp packets and use to check for connection state
	go receiveSimulationTCPAsync()

	err = connectSimulationUDP()
	if err != nil {
		return err
	}
	err = connectClientsUDP()
	if err != nil {
		return err
	}

	// go receiveClientsUDP()
	go receiveSimulationUDPAsync()

	err = runSimulation()
	if err != nil {
		return err
	}

	<-restart

	fmt.Println("Server ended")
	return nil
}
