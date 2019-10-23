package main

import (
	"github.com/fanmanpro/game-server/ws"
)

//var (
//	clients      map[int32]*Client
//	packetQueue  []Packet
//	clientsMutex sync.RWMutex
//)

func main() {
	ip := "127.0.0.1"
	port := "6000"
	wsClient := ws.NewClient(ip, port)
	wsClient.Connect()
}
