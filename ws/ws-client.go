package ws

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"

	"github.com/fanmanpro/game-server/gs"
	"github.com/fanmanpro/game-server/udp"

	"github.com/fanmanpro/coordinator-server/client"
	"github.com/fanmanpro/coordinator-server/gamedata"
	"github.com/fanmanpro/coordinator-server/worker"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var interrupt chan os.Signal
var done chan struct{}
var packetQueue []gamedata.Packet
var connectionID string

// WebSocketClient to open web socket ports
type WebSocketClient struct {
	ip         string
	port       string
	gameServer *gs.GameServer
	udpServer  *udp.UserDatagramProtocolServer
	workerPool *worker.Pool
}

// NewClient initializes a new web socket server without starting it
func NewClient(gameServer *gs.GameServer, ip string, port string, udpServer *udp.UserDatagramProtocolServer) *WebSocketClient {
	return &WebSocketClient{gameServer: gameServer, ip: ip, port: port, udpServer: udpServer, workerPool: worker.NewPool(1, 10)}
}

// Connect establishes connection with websocket server as client
func (w *WebSocketClient) Connect() {
	flag.Parse()
	log.SetFlags(0)

	interrupt = make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: fmt.Sprintf("%v:%v", w.ip, w.port), Path: "/"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	w.workerPool.Start()

	gameServerJoined := &gamedata.GameServerOnline{
		Secret:   "fanmanpro",
		Region:   "Canada",
		Capacity: int32(w.gameServer.Capacity()),
	}
	data, err := ptypes.MarshalAny(gameServerJoined)
	if err != nil {
		log.Printf("err: invalid %v data. err: %v", gamedata.Header_ClientOnline, err)
		return
	}
	packet := gamedata.Packet{
		Header: &gamedata.Header{
			OpCode: gamedata.Header_GameServerOnline,
		},
		Data: data,
	}

	w.workerPool.ScheduleJob(func(errChan chan error) {
		errChan <- w.send(packet, c)
	})

	disconnected := make(chan bool, 1)
	go func() {
		for {
			log.Printf("[WS] waiting for packet")
			mt, data, err := c.ReadMessage()
			if err != nil {
				log.Printf("err: %v", err)
				disconnected <- true
			}
			if mt == websocket.BinaryMessage {
				w.workerPool.ScheduleJob(
					func(errChan chan error) {
						packet := &gamedata.Packet{}
						err = proto.Unmarshal(data, packet)
						if err != nil {
							panic(err)
						}
						log.Printf("receiving packet %v over connection %v from %v to %v", packet.Header.OpCode, packet.Header.Cid, c.RemoteAddr().String(), c.LocalAddr().String())
						log.Printf("recv: %s", packet.Header.OpCode)
						errChan <- w.handlePacket(c, packet)
					},
				)
			}
		}
	}()
	select {
	case <-disconnected:
		{
			fmt.Println("Client disconnected")
		}
		break
	case <-interrupt:
		{
			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "test"))
			if err != nil {
				return
			}
			c.Close()
		}
		break
	}
}

func (w *WebSocketClient) send(p gamedata.Packet, c *websocket.Conn) error {
	log.Printf("sending packet %v over connection %v to %v from %v", p.Header.OpCode, p.Header.Cid, c.RemoteAddr().String(), c.LocalAddr().String())
	data, err := proto.Marshal(&p)
	if err != nil {
		return err
	}

	err = c.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		return err
	}
	log.Printf("sent: %v", p.Header.OpCode)
	return nil
}

func (w *WebSocketClient) handlePacket(c *websocket.Conn, packet *gamedata.Packet) error {
	switch packet.Header.OpCode {
	case gamedata.Header_GameServerOnline:
		{
			connectionID = packet.Header.Cid
			return nil
		}
	case gamedata.Header_GameServerStart:
		{
			if w.udpServer == nil {
				return errors.New("UDP server doesn't exist")
			}

			gameServerStart := &gamedata.GameServerStart{}
			err := ptypes.UnmarshalAny(packet.Data, gameServerStart)
			if err != nil {
				return errors.New(fmt.Sprintf("err: invalid %v data. err: %v", gamedata.Header_GameServerStart, err))
			}
			for _, cl := range gameServerStart.Clients {
				w.gameServer.NewClient(&client.UDPClient{
					Client: &client.Client{
						CID:    cl.ID,
						IPAddr: cl.Address,
					},
					Send: make(chan gamedata.Packet, 1),
				})
			}
			w.gameServer.NewSimulationServer(&client.UDPClient{
				Client: &client.Client{},
				Send:   make(chan gamedata.Packet, 1),
			})
			go w.udpServer.Start()

			//w.gameServer.clients =

			id, err := uuid.NewUUID()
			if err != nil {
				return errors.New(fmt.Sprintf("err: could not generate uuid. %v", err))
			}

			gameServerStart.ID = id.String()

			data, err := ptypes.MarshalAny(gameServerStart)
			if err != nil {
				return errors.New(fmt.Sprintf("err: could not unmarshal packet. %v", err))
			}
			packet := gamedata.Packet{
				Header: &gamedata.Header{
					OpCode: gamedata.Header_GameServerStart,
					Cid:    connectionID,
				},
				Data: data,
			}

			//packetQueue = append(packetQueue, packet)
			w.workerPool.ScheduleJob(func(errChan chan error) {
				errChan <- w.send(packet, c)
			})
			return nil
		}
	default:
		{
			return errors.New(fmt.Sprintf("Packet received but unknown handler for %v", packet.Header.OpCode))
		}
	}
}
