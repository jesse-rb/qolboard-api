package websocket_service

import (
	"net/http"
	"qolboard-api/services/logging"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// Websocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type RoomsManager struct {
	roomsMap  map[uint64]map[*Client]bool
	join      chan *Client
	leave     chan *Client
	broadcast chan RoomMessage
}

type Client struct {
	userUuid string
	canvasId uint64
	conn     *websocket.Conn
	send     chan RoomMessage
}

type RoomMessage struct {
	canvasId uint64
	Event    string `json:"event" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Data     *gin.H `json:"data" binding:"required"`
}

var rm RoomsManager = NewRoomsManager()

func init() {
	go rm.Run()
}

func NewRoomsManager() RoomsManager {
	return RoomsManager{
		roomsMap:  make(map[uint64]map[*Client]bool),
		join:      make(chan *Client),
		leave:     make(chan *Client),
		broadcast: make(chan RoomMessage),
	}
}

func NewClient(userUuid string, canvasId uint64, conn *websocket.Conn) *Client {
	client := &Client{
		userUuid: userUuid,
		canvasId: canvasId,
		send:     make(chan RoomMessage, 256), // Allow for some buffer
		conn:     conn,
	}

	return client
}

func (rm *RoomsManager) Run() {
	// Event loop for our rooms manager, only one of these events should run at a given time
	for {
		select {
		// A client joins a room
		case client := <-rm.join:
			if rm.roomsMap[client.canvasId] == nil {
				rm.roomsMap[client.canvasId] = make(map[*Client]bool)
			}
			rm.roomsMap[client.canvasId][client] = true

		// A client leaves a room
		case client := <-rm.leave:
			if _, ok := rm.roomsMap[client.canvasId][client]; ok {
				delete(rm.roomsMap[client.canvasId], client)
				close(client.send)
			}
			// TODO: Delete room if no more clients? Close ws conn?

		case msg := <-rm.broadcast:
			for c := range rm.roomsMap[msg.canvasId] {
				select {
				// Attempt to send message to the cleint (YAY go channels!)
				case c.send <- msg:
				// client's send queue is full, better not hold up our entire event loop...
				default:
					close(c.send)
					delete(rm.roomsMap[msg.canvasId], c) // bye :)
				}
			}
		}
	}
}

// Allow broadcasting to a room externally
func Broadcast(msg RoomMessage) {
	// Broadcast message to all clients in the room
	rm.broadcast <- msg
}

func Join(userUuid string, canvasId uint64, conn *websocket.Conn) *Client {
	client := NewClient(userUuid, canvasId, conn)
	rm.join <- client

	return client
}

func (c *Client) Leave() {
	rm.leave <- c
	c.conn.Close()
}

func (c *Client) Reader() {
	// Defer handling disconnect
	defer c.Leave()

	for {
		msg := RoomMessage{}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			logging.LogError("WebSocket", "Error reading message from websocket connection", err)
			continue
		}

		// Fwd msg to other clients in the room
		msg.canvasId = c.canvasId
		Broadcast(msg)
	}
}

func (c *Client) Writer() {
	for msg := range c.send {
		err := c.conn.WriteJSON(msg)
		if err != nil {
			logging.LogError("WebSocket", "Error writing message to websocket connection", err)
			continue
		}
	}
}

// WsConnect : Connect client to web socket
func Connect(c *gin.Context) *websocket.Conn {
	// Start websocket connection
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logging.LogError("Connect", "Failed upgrading a connection", err)
	}

	return conn
}
