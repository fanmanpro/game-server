package gs

import (
	"net"

	"github.com/fanmanpro/coordinator-server/gamedata"
)

type GameServer struct {
	clients   []*gamedata.ClientConnection
	addresses []*net.UDPAddr
}

func (g *GameServer) NewClients(clients []*gamedata.ClientConnection) {
	g.clients = clients
	g.addresses = make([]*net.UDPAddr, len(clients))
}

func (g *GameServer) Clients() []*gamedata.ClientConnection {
	return g.clients
}

func (g *GameServer) ClientsConnected() bool {
	for _, c := range g.addresses {
		if c == nil {
			return false
		}
	}
	return true
}

func (g *GameServer) ClientJoined(cid string, c *net.UDPAddr) {
	for i, _ := range g.addresses {
		if g.clients[i].ID == cid {
			g.addresses[i] = c
		}
	}
}
func (g *GameServer) Addresses() []*net.UDPAddr {
	return g.addresses
}

func (g *GameServer) Validate(cid string) bool {
	// when it comes from the sim server
	if cid == "fanmanpro" {
		return true
	}

	for _, c := range g.clients {
		if cid == c.ID {
			return true
		}
	}
	return false
}
