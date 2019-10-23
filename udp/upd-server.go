package udp

// UserDatagramProtocolServer to open web socket ports
type UserDatagramProtocolServer struct {
	ip   string
	port string
}

// New initializes a new web socket server without starting it
func New(ip string, port string) *UserDatagramProtocolServer {
	return &UserDatagramProtocolServer{ip, port}
}

// Start starts the already intialized UDPServer
func (u *UserDatagramProtocolServer) Start() bool {
	return true
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
