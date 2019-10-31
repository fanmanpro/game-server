package udp

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/fanmanpro/coordinator-server/gamedata"
	"github.com/fanmanpro/game-server/gs"
	"github.com/golang/protobuf/proto"
)

var packetQueue []gamedata.Packet

// UserDatagramProtocolServer to open web socket ports
type UserDatagramProtocolServer struct {
	ip         string
	port       string
	gameServer *gs.GameServer
	conn       *net.UDPConn
	address    *net.UDPAddr
	packets    chan gamedata.Packet
}

// New initializes a new web socket server without starting it
func New(gameServer *gs.GameServer, ip string, port string) *UserDatagramProtocolServer {
	return &UserDatagramProtocolServer{gameServer: gameServer, ip: ip, port: port, packets: make(chan gamedata.Packet, 10)}
}

// Start starts the already intialized UDPServer
func (u *UserDatagramProtocolServer) Start() {
	var err error
	u.address, err = net.ResolveUDPAddr("udp4", u.ip+":"+u.port)

	if err != nil {
		log.Fatal(err)
		return
	}

	// setup listener for incoming UDP connection
	conn, err := net.ListenUDP("udp", u.address)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	u.conn = conn

	fmt.Println("UDP server started.")
	connected := make(chan *net.UDPAddr, u.gameServer.Capacity())
	disconnected := make(chan bool, u.gameServer.Capacity())
	simulation := make(chan *net.UDPAddr, 1)

	for i := 0; i < cap(connected); i++ {
		go func() {
			connected <- u.awaitClient(conn)
		}()
		addr := <-connected
		if addr == nil {
			fmt.Println("Client", i, " never joined")
		} else {
			fmt.Println("Client", i, " joined")
			u.gameServer.SetClientAddress(i, addr)
			go func(j int) {
				disconnected <- u.processClient(j, conn, addr)
			}(i)
		}
	}

	fmt.Println("All clients connected")

	go func() {
		simulation <- u.awaitSimulation(conn)
	}()
	simAddress := <-simulation
	fmt.Println("Simulation server connected.")

	go func() {
		u.processSimulation(conn, simAddress)
	}()
	//simulating <- u.processSimulation(j, conn, addr)

	for i := 0; i < cap(disconnected); i++ {
		//	go func() {
		//		disconnected <- u.awaitClient(conn)
		//	}()
		<-disconnected
		fmt.Println("Client", i, " left")
		//	if addr == nil {
		//		fmt.Println("Client", i, " never joined")
		//	} else {
		//		fmt.Println("Client", i, " joined")
		//		u.gameServer.SetClientAddress(i, addr)
		//	}
	}

	fmt.Println("All clients left. Resetting server.")
	//simulating <- false
}
func (u *UserDatagramProtocolServer) processSimulation(conn *net.UDPConn, addr *net.UDPAddr) bool {
	// issues here.
	// - this should be a channel like the rest
	// - cannot send packets to sim server yet
	go func() {
		for {
			buffer := make([]byte, 1024)
			n, _, err := u.conn.ReadFromUDP(buffer)
			if err != nil {
				log.Fatal(err)
			}

			var packet gamedata.Packet
			if err := proto.Unmarshal(buffer[:n], &packet); err != nil { // don't take the full size of the buffer, just the header size
				fmt.Println("Failed to parse address book:", err)
				continue
			}

			//fmt.Println(packet.Header.OpCode)
			switch packet.Header.OpCode {
			case gamedata.Header_SimulationState:
				{
					u.packets <- packet
				}
				break
			}
			//leave <- true
			//break
		}
	}()
	return true
}

func (u *UserDatagramProtocolServer) processClient(i int, conn *net.UDPConn, addr *net.UDPAddr) bool {
	leave := make(chan bool, 1)
	sending := make(chan bool, 1)
	sending <- true
	// reading
	go func() {
		for {
			buffer := make([]byte, 1024)
			n, _, err := u.conn.ReadFromUDP(buffer)
			if err != nil {
				log.Fatal(err)
			}

			var packet gamedata.Packet
			if err := proto.Unmarshal(buffer[:n], &packet); err != nil { // don't take the full size of the buffer, just the header size
				fmt.Println("Failed to parse address book:", err)
				continue
			}
		}
	}()
	go func() {
		for {
			select {
			case packet := <-u.packets:
				{
					out, err := proto.Marshal(&packet)
					if err != nil {
						log.Printf("err: failed to marshal packet. %v", err)
						break
					}
					fmt.Println("Sent packet to", i)
					_, err = u.conn.WriteToUDP(out, addr)
					if err != nil {
						log.Println(err)
					}
				}
				break
			case s := <-sending:
				if !s {
					fmt.Println("Stopped sending to", i)
					break
				}
			}
		}
	}()
	<-leave
	sending <- false
	return true
}

func (u *UserDatagramProtocolServer) awaitClient(conn *net.UDPConn) *net.UDPAddr {
	for {
		buffer := make([]byte, 1024)
		n, addr, err := u.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatal(err)
		}

		var packet gamedata.Packet
		if err := proto.Unmarshal(buffer[:n], &packet); err != nil { // don't take the full size of the buffer, just the header size
			fmt.Println("Failed to parse address book:", err)
			continue
		}

		if packet.Header.OpCode == gamedata.Header_ClientJoin {
			return addr
		}
	}
}

func (u *UserDatagramProtocolServer) awaitSimulation(conn *net.UDPConn) *net.UDPAddr {
	for {
		buffer := make([]byte, 1024)
		n, addr, err := u.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatal(err)
		}

		var packet gamedata.Packet
		if err := proto.Unmarshal(buffer[:n], &packet); err != nil { // don't take the full size of the buffer, just the header size
			fmt.Println("Failed to parse address book:", err)
			continue
		}

		if packet.Header.OpCode == gamedata.Header_SimulationState {
			return addr
		}
	}
}

// Stop closes the UDP connection
func (u *UserDatagramProtocolServer) Stop() {
	u.conn.Close()
}

func (u *UserDatagramProtocolServer) receiving() {
	defer u.Stop()
	for {
		time.Sleep(time.Millisecond)
		buffer := make([]byte, 1024)

		n, _, err := u.conn.ReadFromUDP(buffer)
		if err != nil {
			log.Fatal(err)
		}

		var packet gamedata.Packet
		if err := proto.Unmarshal(buffer[:n], &packet); err != nil { // don't take the full size of the buffer, just the header size
			log.Fatalln("Failed to parse address book:", err)
			return
		}

		//if !u.gameServer.Validate(packet.Header.Cid) {
		//	return
		//}
		u.handlePacket(&packet)
	}
}

func (u *UserDatagramProtocolServer) sending() {
	for {
		time.Sleep(time.Millisecond)
		for _, p := range packetQueue {
			_, err := proto.Marshal(&p)
			if err != nil {
				log.Printf("err: failed to marshal packet. %v", err)
				break
			}

			//out, err := proto.Marshal(&p)
			//if err != nil {
			//	log.Fatalln("Failed to encode address book:", err)
			//}

			//for _, a := range u.gameServer.Addresses() {
			//	if a == nil {
			//		continue
			//	}

			//	_, err = u.conn.WriteToUDP(out, a)
			//	if err != nil {
			//		log.Println(err)
			//	}
			//}
		}
		packetQueue = nil
	}
}
func (u *UserDatagramProtocolServer) handlePacket(packet *gamedata.Packet) {
	switch packet.Header.OpCode {
	case gamedata.Header_SimulationState:
		{
			packetQueue = append(packetQueue, *packet)
		}
		break
	}
}

//service := hostName + ":" + portNum
//udpAddr, err := net.ResolveUDPAddr("udp4", service)

//if err != nil {
//	log.Fatal(err)
//}

//// setup listener for incoming UDP connection
//ln, err := net.ListenUDP("udp", udpAddr)

//if err != nil {
//	log.Fatal(err)
//}

//fmt.Println("UDP server up and listening on port 6000")

//defer ln.Close()

//clients = make(map[int32]*Client)
//// sim server
//clients[0] = nil

//// a client
//clients[9999] = nil
//clientsMutex = sync.RWMutex{}

//go func() {
//	for {
//		writeUDP(ln)
//	}
//}()

//for {
//	readUDP(ln)
//}

//func readUDP(conn *net.UDPConn) {
//	buffer := make([]byte, 1024)
//
//	n, addr, err := conn.ReadFromUDP(buffer)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	var packet Packet
//	if err := proto.Unmarshal(buffer[:n], &packet); err != nil { // don't take the full size of the buffer, just the header size
//		log.Fatalln("Failed to parse address book:", err)
//		return
//	}
//
//	if _, ok := clients[packet.Header.Cid]; !ok {
//		fmt.Println("Invalid Client ID")
//		return
//	}
//
//	clientsMutex.Lock()
//	if clients[packet.Header.Cid] == nil {
//		clients[packet.Header.Cid] = &Client{
//			address: addr,
//		}
//	}
//	clientsMutex.Unlock()
//
//	switch packet.GetHeader().GetOpCode() {
//	case Header_NewGameObject:
//		{
//			// sim server
//			var newGameObject NewGameObject
//			if err := proto.Unmarshal(packet.Data.Value, &newGameObject); err != nil {
//				log.Fatalln("Failed to parse address book:", err)
//				return
//			}
//
//			if packet.Header.Cid == 0 {
//				// came from sim server
//				packetQueue = append(packetQueue, packet)
//			} else {
//				// came from client
//			}
//			//packet.Header.Cid = 0
//			//fmt.Printf("NewGameObject - ID: %v, Prefab: %v CID: %v\n", newGameObject.ID, newGameObject.Prefab, packet.Header.Cid)
//			//clients[packet.Header.Cid].address
//		}
//		break
//	case Header_GameObjectDelta:
//		{
//			// sim server
//			//var vector Vector3
//			//if err := proto.Unmarshal(packet.Data.Value, &vector); err != nil {
//			//	log.Fatalln("Failed to parse address book:", err)
//			//	return
//			//}
//			//clients[packet.Header.Cid].position = vector
//			//packet.Header.Cid = 0
//			//packetQueue = append(packetQueue, packet)
//		}
//		break
//	}
//}
//
//func writeUDP(conn *net.UDPConn) {
//	if packetQueue == nil {
//		return
//	}
//	if clients == nil {
//		return
//	}
//	clientsMutex.Lock()
//	for _, packet := range packetQueue {
//		for cid, client := range clients {
//			if client == nil {
//				continue
//			}
//
//			fmt.Printf("Client: %+v, Packet: %+v\n", cid, packet)
//			if cid == packet.Header.Cid {
//				// it came from this cid, don't send it back...
//				continue
//			}
//
//			out, err := proto.Marshal(&packet)
//			if err != nil {
//				log.Fatalln("Failed to encode address book:", err)
//			}
//
//			_, err = conn.WriteToUDP(out, client.address)
//			if err != nil {
//				log.Println(err)
//			}
//		}
//	}
//	packetQueue = nil
//	clientsMutex.Unlock()
//}

//func repackage(opCode Header_OpCode, m proto.Message) []byte {
//	any, err := ptypes.MarshalAny(m)
//	if err != nil {
//		log.Fatalln("Failed to encode address book:", err)
//	}
//
//	var packet = &Packet{
//		Header: &Header{
//			OpCode: opCode,
//		},
//		Data: any,
//	}
//
//	out, err := proto.Marshal(packet)
//	if err != nil {
//		log.Fatalln("Failed to encode address book:", err)
//	}
//	return out
//}
