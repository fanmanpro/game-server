package main

import (
	"github.com/fanmanpro/coordinator-server/client"
	"github.com/fanmanpro/game-server/gameserver"
	"github.com/fanmanpro/game-server/tcp"
	"github.com/fanmanpro/game-server/udp"
)

func main() {
	tcpServer := tcp.NewServer("127.0.0.1", "1541")
	udpServer := udp.NewServer("127.0.0.1", "1542", "1543")

	gameServer := gameserver.New(udpServer, tcpServer)

	simClient, err := client.New()
	if err != nil {
		return
	}
	gameServer.AddSimulationClient(simClient, "")
	//gameServer.AddSimulationClient(simClient, "/Users/Fanus/source/repos/threaded-network-protocol/Builds/threaded-network-protocol.exe")

	// the coordinator must eventually create new clients when they connection to the WS server and then added to the game server
	{
		client, err := client.NewTest()
		if err != nil {
			return
		}
		gameServer.AddClient(client)
	}

	gameServer.Start()

	//coordinatorIP := "127.0.0.1"
	//coordinatorPort := "1540"
	//wsClient := ws.NewClient(gameServer, coordinatorIP, coordinatorPort, udpServer)
	//wsClient.Connect()
}
