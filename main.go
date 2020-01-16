package main

import (
	"github.com/fanmanpro/coordinator-server/client"
	"github.com/fanmanpro/game-server/gameserver"
	"github.com/fanmanpro/game-server/tcp"
	"github.com/fanmanpro/game-server/udp"
)

func main() {
	clientTCPServer := tcp.NewServer("127.0.0.1", "1541")
	clientUDPServer := udp.NewServer("127.0.0.1", "1542")

	simulationUDPServer := udp.NewServer("127.0.0.1", "1541")

	gameServer := gameserver.New(simulationUDPServer, clientUDPServer, clientTCPServer)

	simClient, err := client.NewSim()
	if err != nil {
		return
	}
	gameServer.AddSimulationClient(simClient, "")
	//gameServer.AddSimulationClient(simClient, "/Users/Fanus/source/repos/threaded-network-protocol/Builds/threaded-network-protocol.exe")

	// the coordinator must eventually create new clients when they connection to the WS server and then added to the game server
	{
		client, err := client.NewExplicit("abc")
		if err != nil {
			return
		}
		gameServer.AddClient(client)
	}

	gameServer.Start()
}
