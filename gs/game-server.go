package gs

import (
	"net"

	"github.com/fanmanpro/game-server/client"
)

type GameServer struct {
	clients []*client.Client
}

func NewGameServer(capacity int) *GameServer {
	return &GameServer{make([]*client.Client, capacity)}
}

func (g *GameServer) NewClient(i int, client *client.Client) {
	g.clients[i] = client
}

func (g *GameServer) SetClientAddress(i int, a *net.UDPAddr) {
	g.clients[i].UDPAddr = a
}

func (g *GameServer) Capacity() int {
	return len(g.clients)
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
