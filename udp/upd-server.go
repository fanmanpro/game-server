package udp

import (
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/fanmanpro/coordinator-server/client"

	"github.com/fanmanpro/coordinator-server/gamedata"
	"github.com/fanmanpro/coordinator-server/worker"
	"github.com/fanmanpro/game-server/gs"
	"github.com/golang/protobuf/proto"
)

var packetQueue []gamedata.Packet

// UserDatagramProtocolServer to open web socket ports
type UserDatagramProtocolServer struct {
	ip         string
	port       string
	gameServer *gs.GameServer
	address    *net.UDPAddr
	workerPool *worker.Pool
}

// New initializes a new web socket server without starting it
func New(gameServer *gs.GameServer, ip string, port string) *UserDatagramProtocolServer {
	return &UserDatagramProtocolServer{gameServer: gameServer, ip: ip, port: port, workerPool: worker.NewPool(10, 100)}
}

func (u *UserDatagramProtocolServer) readUDP(c *net.UDPConn, p gs.GameServerPhase) {
	go func() {
		for {
			//log.Printf("[UDP] waiting for packet")
			buffer := make([]byte, 32768) // also 32KiB
			n, addr, err := c.ReadFromUDP(buffer)
			if err != nil {
				log.Fatal(err)
			}

			u.workerPool.ScheduleJob(
				func(errChan chan error) {
					packet := &gamedata.Packet{}
					err = proto.Unmarshal(buffer[:n], packet)
					if err != nil {
						log.Println(err)
					}
					log.Printf("receiving packet %v over UDP connection %v from origin %v at %v", packet.Header.OpCode, packet.Header.Cid, packet.Header.Origin, addr.String())

					err = u.handlePacket(packet, addr)
					if err != nil {
						log.Println(err)
					}
					errChan <- err
				},
			)
		}
	}()
	<-u.gameServer.PhaseChannel()
}

// perhaps we can run multiple of these is the worker pool to prevent some bottlenecks?
func (u *UserDatagramProtocolServer) startupGameServer(c *net.UDPConn) {
	fmt.Println("UDP server started.", u.ip+":"+u.port)

	u.workerPool.ScheduleJob(func(errChan chan error) {
		errChan <- u.processSimulation(c)
	})
	u.readUDP(c, gs.Start)
	fmt.Println("Simulation server connected")

	fmt.Println(fmt.Sprintf("Preparing for %v clients", u.gameServer.Capacity()))
	u.workerPool.ScheduleJob(func(errChan chan error) {
		errChan <- u.processClients(c)
	})
	go u.readUDP(c, gs.Connect)
	<-u.gameServer.PhaseChannel()
	fmt.Println("Clients connected")

	u.workerPool.ScheduleJob(func(errChan chan error) {
		errChan <- u.processClients(c)
	})
	go u.readUDP(c, gs.Connect)
	<-u.gameServer.PhaseChannel()
}

// Start starts the already intialized UDPServer
func (u *UserDatagramProtocolServer) Start() {
	var err error
	u.address, err = net.ResolveUDPAddr("udp4", u.ip+":"+u.port)

	if err != nil {
		log.Fatal(err)
		return
	}

	// setup listener for incoming UDP connection
	conn, err := net.ListenUDP("udp", u.address)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	u.workerPool.Start()

	u.startupGameServer(conn)

	//connected := make(chan *net.UDPAddr, u.gameServer.Capacity())
	//disconnected := make(chan bool, u.gameServer.Capacity())
	//simulation := make(chan *net.UDPAddr, 1)

	//for i := 0; i < cap(connected); i++ {
	//	go func() {
	//		connected <- u.awaitClient(conn)
	//	}()
	//	addr := <-connected
	//	if addr == nil {
	//		fmt.Println("Client", i, "never joined")
	//	} else {
	//		fmt.Println("Client", i, "joined")
	//		u.gameServer.SetClientAddress(i, addr)
	//		go func(j int) {
	//			disconnected <- u.processClient(j, conn, addr)
	//		}(i)
	//	}
	//}

	//fmt.Println("All clients connected")

	//// once everything is ready and kicks off, start the workers!

	//var simAddress *net.UDPAddr
	//go func() {
	//	simAddress = u.awaitSimulation(conn)
	//	if simAddress != nil {
	//		simulation <- simAddress
	//		return
	//	}
	//}()
	//<-simulation
	//fmt.Println("Simulation server connected.")

	//go func() {
	//	u.processSimulation(conn, simAddress)
	//}()

	//for i := 0; i < cap(disconnected); i++ {
	//	<-disconnected
	//	fmt.Println("Client", i, " left")
	//}

	//fmt.Println("All clients left. Resetting server.")
}
func (u *UserDatagramProtocolServer) handlePacket(packet *gamedata.Packet, addr *net.UDPAddr) error {
	// first check origin
	if packet.Header.Origin {
		// from simulation server
		switch packet.Header.OpCode {
		case gamedata.Header_SimulationJoin: // TODO: Turn into TCP packet
			{
				phase := u.gameServer.Phase()
				if phase != gs.Start {
					return errors.New("Game server not receiving client connections yet")
				}

				packet.Header.OpCode = gamedata.Header_SimulationJoined

				u.gameServer.SetSimulationServerAddress(addr)
				u.gameServer.SendToSimulationServer(packet)
				return nil
			}
		case gamedata.Header_ClientJoined: // TODO: Turn into TCP packet
			{
				u.gameServer.Broadcast(packet)
				return nil
			}
		case gamedata.Header_Velocity:
			{
				u.gameServer.Broadcast(packet)
				return nil
			}
		case gamedata.Header_SimulationState:
			{
				//get it out to all clients
				u.gameServer.Broadcast(packet)
				return nil
			}
		//case gamedata.Header_Rigidbody:
		//	{
		//		out, err := proto.Marshal(packet)
		//		if err != nil {
		//			return err
		//		}
		//		//get it out to all clients
		//		u.gameServer.Broadcast(&out)
		//		return nil
		//	}
		case gamedata.Header_Force:
			{
				//get it out to all clients
				u.gameServer.Broadcast(packet)
				return nil
			}
		}
	} else {
		// from client
		switch packet.Header.OpCode {
		case gamedata.Header_ClientJoin: // TODO: Turn into TCP packet
			{
				phase := u.gameServer.Phase()
				if phase != gs.Connect {
					return errors.New("Game server not receiving client connections yet")
				}

				err := u.gameServer.SetClientAddress(packet.Header.Cid, addr)
				if err != nil {
					return err
				}
				u.gameServer.SendToSimulationServer(packet)
				return nil
			}
		case gamedata.Header_Velocity:
			{
				u.gameServer.SendToSimulationServer(packet)
				return nil
			}
		}
	}
	return errors.New("Invalid packet received from UDP server")
}
func (u *UserDatagramProtocolServer) processSimulation(conn *net.UDPConn) error {
	go func() {
		for {
			select {
			case packet := <-u.gameServer.SimulationServer().Send:
				{
					out, err := proto.Marshal(&packet)
					if err != nil {
						log.Println(err)
						continue
					}

					_, err = conn.WriteToUDP(out, u.gameServer.SimulationServer().UDPAddr)
					if err != nil {
						log.Println(err)
						continue
					}
				}
				break
				//case s := <-sending:
				//	if !s {
				//		fmt.Println("Stopped sending to simulation server")
				//		break
				//	}
			}
		}
	}()
	return nil
}

func (u *UserDatagramProtocolServer) processClients(conn *net.UDPConn) error {
	for _, c := range u.gameServer.Clients() {
		go func(c *client.UDPClient) {
			for {
				select {
				case packet := <-c.Send:
					{
						if c.UDPAddr == nil {
							continue
						}
						out, err := proto.Marshal(&packet)
						if err != nil {
							log.Println(err)
							continue
						}

						_, err = conn.WriteToUDP(out, c.UDPAddr)
						if err != nil {
							log.Println(err)
							continue
						}
						fmt.Println("Sent packet to client", packet.Header.Cid)
					}
					break
				}
			}
		}(c)
	}
	return nil
}
