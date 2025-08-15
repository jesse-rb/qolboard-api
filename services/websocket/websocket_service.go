package websocket_service

import (
	"encoding/json"
	"net/http"
	database_config "qolboard-api/config/database"
	canvas_model "qolboard-api/models/canvas"
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
	author   *Client
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
			logging.LogDebug("(WS event loop)", "receiving join", nil)
			if rm.roomsMap[client.canvasId] == nil {
				rm.roomsMap[client.canvasId] = make(map[*Client]bool)
			}
			rm.roomsMap[client.canvasId][client] = true

		// A client leaves a room
		case client := <-rm.leave:
			logging.LogDebug("(WS event loop)", "receiving leave", nil)
			delete(rm.roomsMap[client.canvasId], client)
			close(client.send)
			client.conn.Close()

			if len(rm.roomsMap[client.canvasId]) <= 0 {
				// If the room is now empty, do some cleanup by deleting the room
				delete(rm.roomsMap, client.canvasId)
			}

		case msg := <-rm.broadcast:
			for c := range rm.roomsMap[msg.canvasId] {
				if msg.author == c {
					continue // Don't send the message back to the author
				}
				select {
				// Attempt to send message to the cleint (YAY go channels!)
				case c.send <- msg:
				// client's send queue is full, better not hold up our entire event loop...
				default:
					logging.LogDebug("(WS event loop)", "skipping... client send channel is FULL", map[string]any{
						"available cap": cap(c.send),
						"queued len":    len(c.send),
					})
					continue
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
}

func (c *Client) Reader(ctx *gin.Context) {
	// Defer handling disconnect
	defer c.Leave()

	for {
		msg := RoomMessage{}
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			logging.LogError("WebSocket", "Error reading message from websocket connection", err)
			break
		}

		shouldUpdateCanvas := msg.Event == "add-piece" || msg.Event == "update-piece" || msg.Event == "update-canvas-data"

		if shouldUpdateCanvas {
			// If needed, save updates to the canvas
			tx, err := database_config.DB(ctx)
			defer tx.Commit()
			if err != nil {
				tx.Rollback()
				break
			}
			canvas, err := canvas_model.Get(tx, c.canvasId)
			var canvasData canvas_model.CanvasData
			err = json.Unmarshal(canvas.CanvasData, &canvasData)
			if err != nil {
				logging.LogError("WebSocket", "Error unmarshalling canvas data", err)
				tx.Rollback()
				break
			}

			if msg.Event == "add-piece" {
				bytes, err := json.Marshal(msg.Data)
				if err != nil {
					logging.LogError("WebSocket", "add-piece -- Could not marhsal piece data", err)
					tx.Rollback()
					break
				}

				var piece canvas_model.PieceData
				err = json.Unmarshal(bytes, &piece)
				if err != nil {
					logging.LogError("WebSocket", "add-piece -- Piece data is invalid", err)
					tx.Rollback()
					break
				}
				canvasData.PiecesManager.Pieces = append(canvasData.PiecesManager.Pieces, &piece)

			} else if msg.Event == "update-piece" {
				bytes, err := json.Marshal(msg.Data)
				if err != nil {
					logging.LogError("WebSocket", "add-piece -- Could not marhsal piece data", err)
					tx.Rollback()
					break
				}

				var piece canvas_model.PieceData
				err = json.Unmarshal(bytes, &piece)
				if err != nil {
					logging.LogError("WebSocket", "add-piece -- Piece data is invalid", err)
					tx.Rollback()
					break
				}

				if index, ok := (*msg.Data)["index"].(int); ok {
					canvasData.PiecesManager.Pieces[index] = &piece
				} else {
					logging.LogError("WebSocket", "update-piece -- Could not find piece by index", map[string]any{
						"msg":      msg,
						"msg.Data": (*msg.Data),
					})
					tx.Rollback()
					break
				}

			} else if msg.Event == "update-canvas-data" {
				bytes, err := json.Marshal((*msg.Data)["canvas_data"])
				if err != nil {
					logging.LogError("WebSocket", "add-piece -- Could not marhsal piece data", err)
					tx.Rollback()
					break
				}

				var incomingCanvasData canvas_model.CanvasData
				err = json.Unmarshal(bytes, &incomingCanvasData)
				if err != nil {
					logging.LogError("WebSocket", "add-piece -- Piece data is invalid", err)
					tx.Rollback()
					break
				}

				canvasData = incomingCanvasData
			}

			canvasDataBytes, err := json.Marshal(canvasData) // TODO: make Canvas.CanvasData the actual CanvasData type
			if err != nil {
				// TODO: Error messages of this and above
				logging.LogError("WebSocket", "Error marshalling canvas data", err)
				tx.Rollback()
				break
			}
			canvas.CanvasData = canvasDataBytes
			logging.LogDebug("WebSocket", "Saving canvas data", nil)
			err = canvas.Save(tx)
			if err != nil {
				logging.LogError("WebSocket", "Error saving canvas data after receiving message", err)
				tx.Rollback()
				break
			}
		}

		// Fwd msg to other clients in the room
		msg.canvasId = c.canvasId
		msg.author = c
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
