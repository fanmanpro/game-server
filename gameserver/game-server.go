package gameserver

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"time"

	"github.com/fanmanpro/coordinator-server/gamedata"
	"github.com/kvz/logstreamer"

	//"github.com/labstack/gommon/log"

	"github.com/fanmanpro/game-server/tcp"
	"github.com/fanmanpro/game-server/udp"

	"github.com/fanmanpro/coordinator-server/client"
)

// Phase is used to keep track of which phase the GameServer is in. Each phase depends on how the clients and simulation process is treated.
type Phase int

const (
	// Initializing phase is when the game server hasn't start yet and not receiving any packets or connections yet
	Initializing Phase = 1
	// Joining phase is when all clients (simulation included) is joining the game server and becoming trusted
	Joining Phase = 2
	// Seating phase is when all clients have joined (or timed out) and the simulation client is ready to start assigning seats
	Seating Phase = 3
	// Running phase is when the simulation is activetly sending and receiving updates to the clients
	Running Phase = 4
	// Concluding is when the simulation is terminated and the game server is getting ready to restart the match/simulation
	Concluding Phase = 5
)

// GameServer TODO
type GameServer struct {
	simUDPServer    *udp.Server
	clientUDPServer *udp.Server
	clientTCPServer *tcp.Server

	phase chan Phase

	clientJoinQueue chan *client.Client
	clientSeatQueue chan *client.Client

	// connected clients to game server
	clients          map[string]*client.Client // use a map?
	simulationClient *client.Client

	// clients connected themselves over the UDP connection
	clientsJoined int
	// simulation server assigned game objects to client seats
	clientsSeated int
}

// New TODO
func New(n int, simUDP *udp.Server, clientUDP *udp.Server, clientTCP *tcp.Server) *GameServer {
	return &GameServer{
		simUDPServer:    simUDP,
		clientUDPServer: clientUDP,
		clientTCPServer: clientTCP,
		phase:           make(chan Phase, 5),                // create a new phase channel when the game server restarts (and push in Initializing phase)
		clients:         make(map[string]*client.Client, n), // capacity should be based on what coordinator tells the game server
		clientJoinQueue: make(chan *client.Client, n),       // capacity should be based on what coordinator tells the game server
		clientSeatQueue: make(chan *client.Client, n),       // capacity should be based on what coordinator tells the game server
	}
}

/*
come up with a plan to test the client and the sim without having to build every time (two unity instances?)
clients are seating properly now. next start testing the udp clients.
make the clients communicate directly via udp but securely over tcp.
freeze the simulation on the first tick (physics), resume it when the game server sends a packet and is in Running state
bring in the context syncing, then the intermitent packets
*/

// Start TODO
func (g *GameServer) Start() {
	g.phase <- Initializing

	// this is out phase manager essentially
	for {
		select {
		case p := <-g.phase:
			{
				switch p {
				case Initializing:
					go g.initializing()
					break
				case Joining:
					go g.joining()
					break
				case Seating:
					go g.seating()
					break
				case Running:
					go g.running()
					break
				case Concluding:
					go g.concluding()
					break
				}
			}
			break
		}
	}
}

func (g *GameServer) initializing() {
	fmt.Printf("[GMS] Phase Start: Initializing\n")
	defer fmt.Printf("[GMS] Phase End: Initializing\n")

	err := g.simUDPServer.Start()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = g.clientUDPServer.Start()
	if err != nil {
		log.Fatal(err)
		return
	}

	err = g.clientTCPServer.Start()
	if err != nil {
		log.Fatal(err)
		return
	}

	g.phase <- Joining
}
func (g *GameServer) joining() {
	// start a timer here to give all players time to connect. if it expires, move to the next phase to start assigning seats
	fmt.Printf("[GMS] Phase Start: Joining\n")
	defer fmt.Printf("[GMS] Phase End: Joining\n")

	g.clientUDPServer.ReadQueue = make(chan *gamedata.Packet, 10)
	g.clientTCPServer.ReadQueue = make(chan *gamedata.Packet, 10)
	g.simUDPServer.ReadQueue = make(chan *gamedata.Packet, 10)

	for {
		select {
		case acceptPacket := <-g.simUDPServer.AcceptQueue:
			{
				err := g.acceptUDPSimClient(acceptPacket.CID, acceptPacket.Addr, acceptPacket.Send)

				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Printf("[GMS] Simulation client trusted\n")

				// because this should be all the player clients, and the simulation client
				if g.clientsJoined == len(g.clients) && g.simulationClient.UDPConn.Addr != nil {
					g.phase <- Seating
					return
				}
			}
		case acceptPacket := <-g.clientUDPServer.AcceptQueue:
			{
				err := g.acceptUDPClient(acceptPacket.CID, acceptPacket.Addr, acceptPacket.Send)

				if err != nil {
					fmt.Println(err)
					continue
				}

				g.clientsJoined++

				fmt.Printf("[GMS] UDP client joined %v\n", g.clientsJoined)

				// because this should be all the player clients, and the simulation client
				if g.clientsJoined == len(g.clients) && g.simulationClient.UDPConn.Addr != nil {
					g.phase <- Seating
					return
				}
			}
		case acceptPacket := <-g.clientTCPServer.AcceptQueue:
			{
				c, err := g.acceptClient(acceptPacket.CID, acceptPacket.Conn, acceptPacket.Send)

				if err != nil {
					fmt.Println(err)
					continue
				}

				// only tell client if they accepted, if they were rejected the connect will time out on the game side
				c.TCPConn.Send <- &gamedata.Packet{
					OpCode: gamedata.Header_ClientTrust,
				}

				fmt.Printf("[GMS] Client trusted\n")

				// obviously we only want to seat the player clients and not the simulation client
				if c != g.simulationClient {
					//g.clientsJoined++
					g.clientSeatQueue <- c
				}
				//if g.clientsJoined == len(g.clients) && g.simulationClient.UDPConn.Addr != nil {
				//	fmt.Println("YO?!1")
				//	g.phase <- Seating
				//	return
				//}
			}
			// we start reading the TCP packets here because clients can already send a disconnect packet that we need to handle
		case p := <-g.clientTCPServer.ReadQueue:
			g.handlePacket(p)
			break
		}
	}
}
func (g *GameServer) seating() {
	// joining is done, start checking that seat queue
	fmt.Printf("[GMS] Phase Start: Seating\n")
	defer fmt.Printf("[GMS] Phase End: Seating\n")

	g.clientTCPServer.ReadQueue = make(chan *gamedata.Packet, 10)
	//g.phase <- Running
	//return
	for {
		select {
		case p := <-g.clientTCPServer.ReadQueue:
			go g.handlePacket(p)
			break
		case p := <-g.clientSeatQueue:
			fmt.Println("Seating client")
			g.clientsSeated++
			g.simulationClient.TCPConn.Send <- &gamedata.Packet{
				OpCode: gamedata.Header_ClientSeat,
				Cid:    p.CID,
			}
			if len(g.clients) == g.clientsSeated {
				g.phase <- Running
				return
			}
			break
		}
	}
}
func (g *GameServer) running() {
	// the simulation is running so now we need to start reading UDP packets
	fmt.Printf("[GMS] Phase Start: Running\n")
	defer fmt.Printf("[GMS] Phase End: Running\n")

	g.clientUDPServer.ReadQueue = make(chan *gamedata.Packet, 10)
	g.clientTCPServer.ReadQueue = make(chan *gamedata.Packet, 10)
	g.simUDPServer.ReadQueue = make(chan *gamedata.Packet, 10)

	for {
		select {
		case p := <-g.clientTCPServer.ReadQueue:
			go g.handlePacket(p)
			continue
		case q := <-g.clientUDPServer.ReadQueue:
			go g.handlePacket(q)
			continue
		case r := <-g.simUDPServer.ReadQueue:
			go g.handlePacket(r)
			continue
		}
	}
}
func (g *GameServer) concluding() {
	fmt.Printf("[GMS] Phase Start: Concluding\n")
	fmt.Printf("[GMS] Phase End: Concluding\n")
	g.phase = make(chan Phase, 5)

	err := g.clientTCPServer.Stop()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("dont forget to also stop the UDP server here")

	g.phase <- Initializing
}

func (g *GameServer) handlePacket(p *gamedata.Packet) {
	switch p.OpCode {
	//
	//TCP Packets
	//
	case gamedata.Header_ClientDisconnect:
		{
		}
		break
	case gamedata.Header_ClientSeat:
		{
			for _, c := range g.clients {
				c.TCPConn.Send <- p
			}
		}
		break
	//
	//UDP Packets
	//
	case gamedata.Header_ClientDatagramAddress:
		{
			//fmt.Println("Header_ClientDatagramAddress")
		}
		break
	case gamedata.Header_Context:
		{
			for _, c := range g.clients {
				c.UDPConn.Send <- p
			}
		}
		break
	}
}

// AddSimulationClient TODO
func (g *GameServer) AddSimulationClient(s *client.Client, p string) error {
	if _, err := os.Stat(p); os.IsNotExist(err) {
		fmt.Printf("[GMS] Not launching simulation client. Expecting editor simulation client. Reason: %v\n", err)
	} else {
		go func() {
			logger := log.New(os.Stdout, "[SIM] ", 0)

			outLogger := logstreamer.NewLogstreamer(logger, "", false)
			defer outLogger.Close()
			errLogger := logstreamer.NewLogstreamer(logger, "", false)
			defer errLogger.Close()

			time.Sleep(1 * time.Second)
			cmd := exec.Command("cmd", "/C", p, "+clientID", s.CID)
			cmd.Stdout = outLogger
			cmd.Stderr = errLogger
			cmd.Start()
		}()

		fmt.Printf("[GMS] -clientID %v\n", s.CID)
	}

	g.simulationClient = s

	return nil
}

// AddClient TODO
func (g *GameServer) AddClient(c *client.Client) {
	g.clients[c.CID] = c
}

func (g *GameServer) acceptUDPSimClient(cid string, addr *net.UDPAddr, send chan *gamedata.Packet) error {
	if g.simulationClient.Addr.Equal(addr.IP) && (g.simulationClient.CID == cid) {
		g.simulationClient.UDPConn.Addr = addr
		g.simulationClient.UDPConn.Send = send
		fmt.Printf("[GMS] Simulation client address accepted. %v \n", addr.String())
		return nil
	}
	return fmt.Errorf("[GMS] Simulation client address rejected. %v", addr.String())
}

func (g *GameServer) acceptUDPClient(cid string, addr *net.UDPAddr, send chan *gamedata.Packet) error {
	if cid == "" {
		fmt.Printf("[GMS] Client rejected. Empty CID given.\n")
	}
	for _, c := range g.clients {
		if c.Addr.Equal(addr.IP) && c.CID == cid {
			c.UDPConn.Addr = addr
			c.UDPConn.Send = send
			fmt.Printf("[GMS] Client address accepted. %v\n", addr.String())
			return nil
		}
	}
	return fmt.Errorf("[GMS] Client address rejected. %v", addr.String())
}

func (g *GameServer) acceptClient(cid string, conn *net.TCPConn, send chan *gamedata.Packet) (*client.Client, error) {
	remoteAddr := (conn.RemoteAddr().(*net.TCPAddr))
	localAddr := (conn.LocalAddr().(*net.TCPAddr))
	// the simulation client has to be local to the game server and have a specific port
	if remoteAddr.Port == 9737 && remoteAddr.IP.Equal(localAddr.IP) {
		g.simulationClient.Addr = remoteAddr.IP
		g.simulationClient.TCPConn.Conn = conn
		g.simulationClient.TCPConn.Send = send
		fmt.Printf("[GMS] Simulation client accepted. %v %v\n", cid, g.simulationClient.Addr.String())
		return g.simulationClient, nil
	}

	if c, ok := g.clients[cid]; ok {
		c.Addr = remoteAddr.IP
		fmt.Printf("[GMS] Client accepted. %v %v\n", cid, c.Addr.String())
		c.TCPConn.Conn = conn
		c.TCPConn.Send = send
		return c, nil
	}
	return nil, fmt.Errorf("[GMS] Client rejected. %v", cid)
}

// Clients TODO
func (g *GameServer) Clients() map[string]*client.Client {
	return g.clients
}

// Client TODO
func (g *GameServer) Client(c string) *client.Client {
	return g.clients[c]
}

// SimulationClient TODO
func (g *GameServer) SimulationClient() *client.Client {
	return g.simulationClient
}
