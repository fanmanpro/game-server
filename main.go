package main

import (
	"github.com/fanmanpro/game-server/gs"
	"github.com/fanmanpro/game-server/udp"
	"github.com/fanmanpro/game-server/ws"
)

func main() {
	gameServer := gs.NewGameServer(2)

	localIP := "127.0.0.1"
	localPort := "1541"
	udpServer := udp.New(gameServer, localIP, localPort)

	coordinatorIP := "127.0.0.1"
	coordinatorPort := "1540"
	wsClient := ws.NewClient(gameServer, coordinatorIP, coordinatorPort, udpServer)
	wsClient.Connect()
}
