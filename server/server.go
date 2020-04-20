package server

import (
	"errors"
	"fmt"
	"io"
	"log"
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

// type Server struct {
// }

// func NewServer() *Server {
// 	return &Server{}
// 	// return &Server{rate: 50, tick: 1, stop: make(chan bool)} // 100 is 10 ticks per second, 50 is 20, 33 is 30, etc.
// }

var simulationDisconnected bool
var simulationTCPConnection *net.TCPConn
var simulationTCPAddressHost *net.TCPAddr
var simulationTCPAddressRemote *net.TCPAddr
var simulationUDPConnection *net.UDPConn
var simulationUDPAddressHost *net.UDPAddr
var simulationUDPAddressRemote *net.UDPAddr

var clientTCPConnections map[MAC]*net.TCPConn
var clientTCPAddressHost *net.TCPAddr
var clientUDPConnections map[*net.TCPConn]*net.UDPConn
var clientUDPAddressHost *net.UDPAddr

// GUID is an identifier for a netsync object
type GUID = string

// MAC is an unique network adapter identifier
type MAC = string

var seats map[GUID]*MAC

// var connectionsTCP map[MAC]*net.TCPConn
// var connectionsUDP map[MAC]*net.TCPConn

// var connections map[MAC]*net.TCPConn

var macCh chan MAC

// var connectionCh chan *net.TCPConn

const rate time.Duration = 50

var tick int32 = 1

var sent time.Time
var end time.Time

func connectSimulationTCP() error {
	var err error

	// simulationTCPAddressHost, err = net.ResolveTCPAddr(tcpNetwork, fmt.Sprintf("35.183.5.196:%v", "9999"))
	simulationTCPAddressHost, err = net.ResolveTCPAddr(tcpNetwork, fmt.Sprintf(":%v", "9999"))
	// simulationTCPAddressHost, err = net.ResolveTCPAddr(tcpNetwork, fmt.Sprintf("%v:%v", localhostAddress, "9999"))
	if err != nil {
		return err
	}

	simulationTCPAddressRemote, err = net.ResolveTCPAddr(tcpNetwork, fmt.Sprintf(":%v", "1999"))
	if err != nil {
		return err
	}

	fmt.Printf("simulation connecting from %v to %v\n", simulationTCPAddressHost.String(), simulationTCPAddressRemote.String())
	// connect the sim
	for {
		// simulationTCPConnection, err = net.ListenTCP("tcp4"u tcpNetwork) //, simulationTCPAddressHost, simulationTCPAddressRemote)
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

	clientTCPAddressHost, err = net.ResolveTCPAddr(tcpNetwork, fmt.Sprintf(":%v", "9889"))
	if err != nil {
		fmt.Println(err)
		return
	}

	// this should happen per client

	// split the listening and the seating
	// because we need to always listen and only push into a seating channel
	// this will allow us to connect into a new seat, or reconnect into a disconnected seat through a ip & mac address combo

	/*
		running sim, reconnecting client journey
		✓ tcp connect
		✓ receive mac address
		✓ A store { seatguid: mac } map
		✓ B store { mac: tcpconn } map immediately after above
		- client disconnects
		- stop reading and write tcp + udp
		- client reconnects
		- client mac was found in B and client ip is compared with old tcpconn ip
		- success: reconnect and seat at A. start reading and writing tcp + udp
		- failed: look for open seats just like during connect

		running clients, reconnecting sim journey
		- tcp connect
		- configure seats
		- sim disconnects
		- stop reading and writing tcp + udp
		- sim reconnects
		- check B map for existing mac + ip entries
		- A store for each entry in B
	*/
	var clientTCPListener *net.TCPListener
	clientTCPListener, err = net.ListenTCP("tcp", clientTCPAddressHost)
	if err != nil {
		fmt.Println(err)
		return
	}
	// clientTCPListener.SetDeadline(time.Now().Add(1 * time.Second))

	for {
		clientTCPConnection, err := clientTCPListener.AcceptTCP()
		fmt.Println("client attempting to (re)connect")
		if err != nil {
			fmt.Println(err)
			// fmt.Println("only", i, "client(s) managed to connected: ", err)
			// seatsConn <- nil
			return
		}
		go receiveClientMacTCPAsync(clientTCPConnection)
	}

	// seatCount := len(seats)
	// for i := 0; i < seatCount; i++ {
	// 	fmt.Println("client connected")
	// 	clientTCPConnections = append(clientTCPConnections, clientTCPConnection)
	// 	seatsConn <- clientTCPConnection
	// 	go receiveClientTCP(clientTCPConnection)
	// }
	// return
}
func receiveClientMacTCPAsync(clientTCPConnection *net.TCPConn) {
	var err error
	var l int

	// wait for context
	clientTCPConnection.SetReadDeadline(time.Now().Add(time.Second * 5))

	buffer := make([]byte, 64*1024*1024)
	l, err = clientTCPConnection.Read(buffer)
	if err != nil {
		if err == io.EOF {
			fmt.Println("client disconnected while awaiting its MAC")
		}
		if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
			fmt.Println("client timed out while awaiting its MAC")
		}
		return
	}

	// unmarshal context from client. validate then forward.
	data := buffer[:l]
	packet := &serializable.Packet{}
	err = proto.Unmarshal(data, packet)
	if err != nil {
		log.Println(err)
		return
	}

	if packet.OpCode == serializable.Packet_ClientMAC {
		clientMac := &serializable.ClientMAC{}
		err = ptypes.UnmarshalAny(packet.Data, clientMac)
		if err != nil {
			log.Println(err)
			return
		}

		if oldClientTCPConnection, ok := clientTCPConnections[clientMac.MAC]; ok {
			fmt.Println("client reconnecting with mac. checking address")
			oldIP := strings.Split(oldClientTCPConnection.RemoteAddr().String(), ":")[0]
			newIP := strings.Split(clientTCPConnection.RemoteAddr().String(), ":")[0]
			// fmt.Println(oldClientTCPConnection.RemoteAddr().String(), clientTCPConnection.RemoteAddr().String())
			if oldIP == newIP {
				fmt.Println("client reconnect matched")
				clientTCPConnections[clientMac.MAC] = clientTCPConnection
				for guid, mac := range seats {
					if *mac == clientMac.MAC {
						sendClientSeat(clientTCPConnection, guid)
					}
				}
				go receiveClientTCPAsync(clientTCPConnection)

				err = connectClientUDP(clientTCPConnection)
				if err != nil {
					log.Println(err)
				}
				// successful reconnect
				return
			}
		}

		// not a reconnect so new client and new mac
		clientTCPConnections[clientMac.MAC] = clientTCPConnection
		macCh <- clientMac.MAC
		go receiveClientTCPAsync(clientTCPConnection)
	}
}

// this fillSeats isn't really dynamic, it needs to rather fill
// them one at a time continuously as they become available to
// allow of any kind of game server seat filling possibility like
// BR (fill until full), MMORPG (always fill), War3 (fill on request)
func fillSeats() error {
	for guid := range seats {
		mac := <-macCh
		fmt.Printf("seating at %v for %v\n", guid, mac)
		seats[guid] = &mac
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
		seatCount := len(seatConfiguration.Seats)
		fmt.Printf("server registered %v seats from simulation\n", seatCount)

		// later on when server has a growing seat cat, remove the map caps below
		seats = make(map[GUID]*MAC, seatCount)
		clientTCPConnections = make(map[MAC]*net.TCPConn, seatCount)
		clientUDPConnections = make(map[*net.TCPConn]*net.UDPConn, seatCount)
		macCh = make(chan MAC)
		for _, seat := range seatConfiguration.Seats {
			fmt.Printf("seating available at %v\n", seat.GUID)
			seats[seat.GUID] = nil
		}
	} else {
		return errors.New("invalid seat configuration packet opcode")
	}
	return nil
}

func connectClientUDP(clientTCPConnection *net.TCPConn) error {
	var err error
	var clientUDPAddressRemote *net.UDPAddr

	clientUDPAddressRemote, err = net.ResolveUDPAddr(udpNetwork, fmt.Sprintf("%v:%v", strings.Split(clientTCPConnection.RemoteAddr().String(), ":")[0], "1888")) // this should be a client address, not localhost
	if err != nil {
		return err
	}

	clientUDPConnection, err := net.DialUDP(udpNetwork, clientUDPAddressHost, clientUDPAddressRemote)
	if err != nil {
		return err
	}

	clientUDPConnections[clientTCPConnection] = clientUDPConnection
	go receiveClientUDP(clientUDPConnection)
	return nil
}

func connectClientsUDP() error {
	var err error

	clientUDPAddressHost, err = net.ResolveUDPAddr(udpNetwork, fmt.Sprintf(":%v", "9888")) // this should be a client address, not localhost
	if err != nil {
		return err
	}

	for _, clientTCPConnection := range clientTCPConnections {
		if clientTCPConnection == nil {
			continue
		}

		err = connectClientUDP(clientTCPConnection)
		if err != nil {
			return err
		}
	}

	return nil
}

func connectSimulationUDP() error {
	simulationUDPAddressHost, err := net.ResolveUDPAddr(udpNetwork, fmt.Sprintf(":%v", "9998"))
	if err != nil {
		return err
	}

	simulationUDPAddressRemote, err := net.ResolveUDPAddr(udpNetwork, fmt.Sprintf(":%v", "1998"))
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

func sendClientSeat(clientTCPConnection *net.TCPConn, guid GUID) error {
	ss := &serializable.Seat{
		GUID: guid,
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

	_, err = clientTCPConnection.Write(data)
	if err != nil {
		if err.(net.Error).Timeout() {
			fmt.Printf("client disconnected 2\n")
			return nil
		}
		return err
	}
	return nil
}

func broadcastSeats() error {
	for m, c := range clientTCPConnections {
		if c == nil {
			fmt.Println("address was nil")
			continue
		}
		var guid GUID
		for g, mac := range seats {
			if *mac == m {
				guid = g
			}
		}
		if guid == "" {
			fmt.Println("connection had no seat associated")
			continue
		}

		// fmt.Printf("%v seated at %v\n", a.Port, g)

		err := sendClientSeat(c, guid)
		if err != nil {
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
func receiveClientTCPAsync(clientTCPConnection *net.TCPConn) {
	if clientTCPConnection == nil {
		fmt.Println("no TCP connection for client to receive with")
	} else {
		fmt.Println("receiving TCP for client")
	}
	defer disconnectClient(clientTCPConnection)
	clientTCPConnection.SetReadDeadline(time.Time{})
	for {
		// wait for context
		buffer := make([]byte, 64*1024*1024)
		_, err := clientTCPConnection.Read(buffer)
		if err != nil {
			if err == io.EOF {
				fmt.Println("client disconnected")
				return
			}
			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				fmt.Println("client timed out")
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

func disconnectClient(clientTCPConnection *net.TCPConn) {
	clientTCPConnection.SetLinger(0)
	clientTCPConnection.SetKeepAlive(false)
	clientTCPConnection.Close()

	clientUDPConnections[clientTCPConnection].Close()
	delete(clientUDPConnections, clientTCPConnection)

	fmt.Println("client connections closed")
	// fmt.Println(clientTCPConnection.RemoteAddr().String())

	// start reconnect timer and dispose connections if timer runs out

	// reconnectClient()
}

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

	// connectClientsTCPAsync()
}

func Start() error {
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
