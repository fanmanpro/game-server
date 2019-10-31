package ws

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/fanmanpro/game-server/client"
	"github.com/fanmanpro/game-server/gs"
	"github.com/fanmanpro/game-server/udp"

	"github.com/fanmanpro/coordinator-server/gamedata"
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
	conn       *websocket.Conn
	gameServer *gs.GameServer
	udpServer  *udp.UserDatagramProtocolServer
}

// NewClient initializes a new web socket server without starting it
func NewClient(gameServer *gs.GameServer, ip string, port string, udpServer *udp.UserDatagramProtocolServer) *WebSocketClient {
	return &WebSocketClient{gameServer: gameServer, ip: ip, port: port, udpServer: udpServer}
}

// Disconnect closes connection to web socket server
func (w *WebSocketClient) Disconnect() {
	w.conn.Close()
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
	w.conn = c
	defer w.Disconnect()

	done = make(chan struct{})
	go w.receiving()

	gameServerJoined := &gamedata.GameServerOnline{
		Secret:   "fanmanpro",
		Region:   "Canada",
		Capacity: 1,
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

	packetQueue = append(packetQueue, packet)

	w.sending()
}

func (w *WebSocketClient) receiving() {
	defer close(done)
	for {
		time.Sleep(time.Millisecond)
		mt, message, err := w.conn.ReadMessage()
		if err != nil {
			log.Printf("err: %v", err)
			break
		}
		if mt == websocket.BinaryMessage {
			packet := &gamedata.Packet{}
			err = proto.Unmarshal(message, packet)
			if err != nil {
				panic(err)
			}
			log.Printf("recv: %s", packet.Header.OpCode)
			w.handlePacket(w.conn, packet)
		}
	}
}

func (w *WebSocketClient) sending() {
	for {
		time.Sleep(time.Millisecond)
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := w.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
				return
			}
		default:
			{
				for _, p := range packetQueue {
					data, err := proto.Marshal(&p)
					if err != nil {
						log.Printf("err: failed to marshal packet. %v", err)
						break
					}

					err = w.conn.WriteMessage(websocket.BinaryMessage, data)
					if err != nil {
						log.Println("err:", err)
						break
					} else {
						log.Printf("sent: %v", p.Header.OpCode)
					}
				}
				packetQueue = nil
			}
		}
	}
}

func (w *WebSocketClient) handlePacket(c *websocket.Conn, packet *gamedata.Packet) {
	switch packet.Header.OpCode {
	case gamedata.Header_GameServerOnline:
		{
			connectionID = packet.Header.Cid
		}
		break
	case gamedata.Header_GameServerStart:
		{
			if w.udpServer == nil {
				log.Printf("err: udp server doesn't exist")
				return
			}

			gameServerStart := &gamedata.GameServerStart{}
			err := ptypes.UnmarshalAny(packet.Data, gameServerStart)
			if err != nil {
				log.Printf("err: invalid %v data. err: %v", gamedata.Header_GameServerStart, err)
				return
			}
			for i, cl := range gameServerStart.Clients {
				w.gameServer.NewClient(i, &client.Client{
					CID:   cl.ID,
					IPddr: cl.Address,
				})
			}
			go w.udpServer.Start()
			//w.gameServer.clients =

			id, err := uuid.NewUUID()
			if err != nil {
				log.Printf("err: could not generate uuid. %v", err)
				break
			}

			gameServerStart.ID = id.String()

			data, err := ptypes.MarshalAny(gameServerStart)
			if err != nil {
				log.Printf("err: could not generate uuid. %v", err)
				return
			}
			packet := gamedata.Packet{
				Header: &gamedata.Header{
					OpCode: gamedata.Header_GameServerStart,
					Cid:    connectionID,
				},
				Data: data,
			}

			packetQueue = append(packetQueue, packet)
		}
		break
	}
}
