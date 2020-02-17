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

	gameServer := gameserver.New(2, simulationUDPServer, clientUDPServer, clientTCPServer)

	/*
		ok so do this, make the go server tick at a rate. on each tick, ask and receive the context from the sim server
		once received, send to all clients.
		when client sends a owned object packet, (validate it then) queue it with all the other client owned packets for the next tick
		the client owned packet contains transform data, as well as all the actions the simulation must respond to. e.g. spawning of
		player owned objects

		for the sim server, when it connects, send all the game objects guids to the go server, and which seat owns which. this
		allows the go server to take care of the seating. once all the seats are full, send the first UDP packet, which will start
		the simulation.

		this approach is good for a few reasons:
		- validating whether a client sent packet for a game object, is done on the go server
		- only once all the seats are fill with the simulation ACTUALLY start simulating.
		- clients receive states, not deltas
		- the simulation has no bottlenecks and the go server dictates the tick rate based on when it wants to read the sim


		TASK:
		write the pseudo code from starting each server to spawning a client owned gameobject on the simulation

		tcp server starts
		tcp awaits to accept sim tcp client
		tcp listen for seats packet
		* handle seats packet, assigning seats, and all the GUIDs that are owned by that seat
		tcp listen for seat packet
		* handle seat packet, filling each seat with a tcp and udp address
		-> client starts udp listen
		-> clients load scene (always pre-paused)
		countdown before ticks start... allows clients to load scene
		udp read for context packet
		udp write for context packet
		update start
		> udp send context to sim (unpause if paused)
		> udp receive context from sim
		> udp send context to each client (unpause if paused)
		> udp receive context from each client
		> verify if context tick is current tick. if not, register packet loss
		> verify if GUIDs are owned by client (just do this once a second?)
		> combine context of each client
		> upd send context to sim
		update end



	*/

	simClient, err := client.NewSim()
	if err != nil {
		return
	}
	gameServer.AddSimulationClient(simClient, "")
	//gameServer.AddSimulationClient(simClient, "/Users/Fanus/source/repos/threaded-network-protocol/Builds/threaded-network-protocol.exe")

	// the coordinator must eventually create new clients when they connection to the WS server and then added to the game server
	{
		client1, err := client.NewExplicit("abc")
		if err != nil {
			return
		}
		gameServer.AddClient(client1)

		client2, err := client.NewExplicit("def")
		if err != nil {
			return
		}
		gameServer.AddClient(client2)
	}

	gameServer.Start()
}
