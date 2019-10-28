package ws

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

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
	ip   string
	port string
}

// NewClient initializes a new web socket server without starting it
func NewClient(ip string, port string) *WebSocketClient {
	return &WebSocketClient{ip, port}
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
	defer c.Close()

	done = make(chan struct{})
	go receiving(c)

	//go func() {
	//	defer close(done)
	//	for {
	//		_, message, err := c.ReadMessage()
	//		if err != nil {
	//			log.Println("read:", err)
	//			return
	//		}
	//		log.Printf("recv: %s", message)
	//	}
	//}()

	//ticker := time.NewTicker(time.Second)
	//defer ticker.Stop()
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

	sending(c)

	//for {
	//	select {
	//	case <-done:
	//		return
	//	case t := <-ticker.C:
	//		fmt.Println("here")
	//		err := c.WriteMessage(websocket.BinaryMessage, []byte(t.String()))
	//		if err != nil {
	//			log.Println("write:", err)
	//			return
	//		}
	//	case <-interrupt:
	//		log.Println("interrupt")

	//		// Cleanly close the connection by sending a close message and then
	//		// waiting (with timeout) for the server to close the connection.
	//		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	//		if err != nil {
	//			log.Println("write close:", err)
	//			return
	//		}
	//		select {
	//		case <-done:
	//		case <-time.After(time.Second):
	//		}
	//		return
	//	}
	//}
}

func receiving(c *websocket.Conn) {
	defer close(done)
	for {
		time.Sleep(time.Millisecond)
		mt, message, err := c.ReadMessage()
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
			handlePacket(c, packet)
		}
	}
}

func sending(c *websocket.Conn) {
	for {
		time.Sleep(time.Millisecond)
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
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

					err = c.WriteMessage(websocket.BinaryMessage, data)
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

func handlePacket(c *websocket.Conn, packet *gamedata.Packet) {
	switch packet.Header.OpCode {
	case gamedata.Header_GameServerOnline:
		{
			connectionID = packet.Header.Cid

			fmt.Println(connectionID)
			//gameServerJoined := &gamedata.GameServerJoined{
			//	Id:     "fanmanpro",
			//	Region: "Canada",
			//}
			//err := ptypes.UnmarshalAny(data, gameServerJoined)
			//if err != nil {
			//	log.Printf("err: invalid %v data. err: %v", gamedata.Header_ClientJoined, err)
			//	return
			//}

			// get the player info from the database
			//clientJoined.Name = "FanManPro"

			//data, err = ptypes.MarshalAny(clientJoined)
			//if err != nil {
			//	log.Printf("err: could not generate uuid. %v", err)
			//	return
			//}
			//packet := gamedata.Packet{
			//	Header: &gamedata.Header{
			//		OpCode: opCode,
			//		Cid:    id.String(),
			//	},
			//	Data: data,
			//}

			//packetQueue = append(packetQueue, packet)
		}
		break
	case gamedata.Header_GameServerStart:
		{
			fmt.Println(connectionID)
			gameServerStart := &gamedata.GameServerStart{}
			err := ptypes.UnmarshalAny(packet.Data, gameServerStart)
			if err != nil {
				log.Printf("err: invalid %v data. err: %v", gamedata.Header_GameServerStart, err)
				return
			}

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
