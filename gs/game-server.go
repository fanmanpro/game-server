package gs

import (
	"errors"
	"fmt"
	"net"

	"github.com/fanmanpro/coordinator-server/gamedata"

	"github.com/fanmanpro/coordinator-server/client"
)

type GameServerPhase int

const (
	Start   GameServerPhase = 0
	Connect GameServerPhase = 1
	Assign  GameServerPhase = 2
	Active  GameServerPhase = 3
	End     GameServerPhase = 4
)

type GameServer struct {
	// channel to manage game server's progression through phases
	phase        GameServerPhase
	phaseChannel chan GameServerPhase

	// channel to manage client's connection during Connect phase
	clientChannel chan *client.UDPClient

	// connected clients to game server
	clients          []*client.UDPClient
	simulationServer *client.UDPClient

	// clients were created by coordinators request
	clientsJoined int
	// clients connected themselves over the UDP connection
	clientsConnected int
	// simulation server assigned game objects to client seats
	clientsAssigned int
}

func NewGameServer(capacity int) *GameServer {
	return &GameServer{phaseChannel: make(chan GameServerPhase, 1), clientChannel: make(chan *client.UDPClient, 1), clients: make([]*client.UDPClient, 2)}
}

func (g *GameServer) NewSimulationServer(simulationServer *client.UDPClient) {
	g.simulationServer = simulationServer
}

func (g *GameServer) SetSimulationServerAddress(a *net.UDPAddr) {
	g.simulationServer.UDPAddr = a
	g.phaseChannel <- 1
	g.phase = 1
}

func (g *GameServer) NewClient(client *client.UDPClient) {
	g.clients[g.clientsJoined] = client
	g.clientsJoined++
}

func (g *GameServer) SetClientAddress(cid string, a *net.UDPAddr) error {
	for i, c := range g.clients {
		if c.Client.CID == cid {
			if g.clients[i].UDPAddr == nil {
				g.clients[i].UDPAddr = a
				g.clientsConnected++
				if 2 == g.clientsConnected {
					g.phaseChannel <- 2
					g.phase = 2
				}
				return nil
			}
		}
	}
	for _, c := range g.clients {
		fmt.Printf("%+v\n", c.Client.CID)
	}

	return errors.New(fmt.Sprintf("this is weird. some unknown client trying to join. %v", cid))
}

func (g *GameServer) Capacity() int {
	return 2
}

func (g *GameServer) Phase() GameServerPhase {
	return g.phase
}

func (g *GameServer) PhaseChannel() chan GameServerPhase {
	return g.phaseChannel
}

func (g *GameServer) Clients() []*client.UDPClient {
	return g.clients
}

func (g *GameServer) ClientAt(i int) *client.UDPClient {
	return g.clients[i]
}

func (g *GameServer) SimulationServer() *client.UDPClient {
	return g.simulationServer
}

func (g *GameServer) Broadcast(packet *gamedata.Packet) {
	for _, c := range g.clients {
		p := *packet
		p.Header.Cid = c.Client.CID
		c.Send <- p
	}
}

func (g *GameServer) SendToSimulationServer(packet *gamedata.Packet) {
	g.simulationServer.Send <- *packet
}

//func (g *GameServer) Clients() []*gamedata.ClientConnection {
//	return g.clients
//}
//
//func (g *GameServer) Addresses() []*net.UDPAddr {
//	return g.addresses
//}
//
//func (g *GameServer) ClientsConnected() bool {
//	for _, c := range g.addresses {
//		if c == nil {
//			return false
//		}
//	}
//	return true
//}

//func (g *GameServer) ClientJoined(cid string, c *net.UDPAddr) {
//	for i, _ := range g.addresses {
//		if g.clients[i].ID == cid {
//			g.addresses[i] = c
//		}
//	}
//}
//
//func (g *GameServer) Validate(cid string) bool {
//	// when it comes from the sim server
//	if cid == "fanmanpro" {
//		return true
//	}
//
//	for _, c := range g.clients {
//		if cid == c.ID {
//			return true
//		}
//	}
//	return false
//}
