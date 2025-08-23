package websocket_service

import (
	"encoding/json"
	"net/http"
	database_config "qolboard-api/config/database"
	model "qolboard-api/models"
	canvas_model "qolboard-api/models/canvas"
	service "qolboard-api/services"
	"qolboard-api/services/logging"
	response_service "qolboard-api/services/response"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jesse-rb/imissphp-go"
)

// Websocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type joinRoomData struct {
	userUuid string
	canvas   *model.Canvas
	conn     *websocket.Conn
	chResume chan *Client
}

type RoomsManager struct {
	roomsMap  map[uint64]*Room
	join      chan *joinRoomData
	leave     chan *Client
	broadcast chan RoomMessage
}

type Room struct {
	Canvas  *model.Canvas
	Clients map[*Client]bool
}

type Client struct {
	userUuid string
	room     *Room
	conn     *websocket.Conn
	send     chan RoomMessage
}

type RoomMessage struct {
	author *Client
	room   *Room
	Event  string         `json:"event" binding:"required"`
	Email  string         `json:"email" binding:"required"`
	Data   map[string]any `json:"data" binding:"required"`
}

var rm RoomsManager = NewRoomsManager()

func init() {
	go rm.Run()
}

func NewRoomsManager() RoomsManager {
	return RoomsManager{
		roomsMap:  make(map[uint64]*Room),
		join:      make(chan *joinRoomData),
		leave:     make(chan *Client),
		broadcast: make(chan RoomMessage),
	}
}

func NewRoom(canvas *model.Canvas) *Room {
	return &Room{
		Canvas:  canvas,
		Clients: make(map[*Client]bool),
	}
}

func NewClient(userUuid string, room *Room, conn *websocket.Conn) *Client {
	client := &Client{
		userUuid: userUuid,
		room:     room,
		send:     make(chan RoomMessage, 256), // Allow for some buffer
		conn:     conn,
	}

	return client
}

func BroadcastCanvas(canvas *model.Canvas) {
	// data := gin.H(service.ToMapStringAny(canvas))
	data := response_service.BuildResponse(*canvas)
	if v, ok := data.(map[string]any); ok {
		msg := RoomMessage{
			author: nil,
			room:   rm.getRoom(canvas),
			Event:  "get",
			Email:  "",
			Data:   v,
		}

		Broadcast(msg)
	}
}

func (r *Room) addClient(client *Client) {
	r.Clients[client] = true
}

func (r *Room) removeClient(client *Client) {
	delete(r.Clients, client)
}

func (r *Room) hasClients() bool {
	return len(r.Clients) > 0
}

func (rm *RoomsManager) getRoom(canvas *model.Canvas) *Room {
	if room, exists := rm.roomsMap[canvas.ID]; exists {
		return room
	}
	room := NewRoom(canvas)
	rm.roomsMap[canvas.ID] = room

	return room
}

func (rm *RoomsManager) Run() {
	// Event loop for our rooms manager, only one of these events should run at a given time
	for {
		select {
		// A client joins a room
		case joinRoomData := <-rm.join:
			logging.LogDebug("(WS event loop)", "receiving join", nil)
			room := rm.getRoom(joinRoomData.canvas)
			client := NewClient(joinRoomData.userUuid, room, joinRoomData.conn)
			room.addClient(client)
			joinRoomData.chResume <- client // Send the client back to the websocket connection controller action

		// A client leaves a room
		case client := <-rm.leave:
			logging.LogDebug("(WS event loop)", "receiving leave", nil)
			room := client.room
			room.removeClient(client)
			close(client.send)
			client.conn.Close()

			if !room.hasClients() {
				// If the room is now empty, do some cleanup by deleting the room
				delete(rm.roomsMap, room.Canvas.ID)
			}

		case msg := <-rm.broadcast:
			for c := range msg.room.Clients {
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

func Join(userUuid string, canvas *model.Canvas, conn *websocket.Conn, chResume chan *Client) {
	rm.join <- &joinRoomData{
		userUuid: userUuid,
		canvas:   canvas,
		conn:     conn,
		chResume: chResume,
	}
}

func (c *Client) Leave() {
	rm.leave <- c
}

func (c *Client) Reader(ctx *gin.Context) {
	// Defer handling disconnect
	defer c.Leave()
	logging.LogDebug("WebSocket", "Reader -- Starting reader for client", nil)
	room := c.room
	canvas := room.Canvas

	for {
		msgIncoming := RoomMessage{}
		err := c.conn.ReadJSON(&msgIncoming)
		if err != nil {
			logging.LogError("WebSocket", "Error reading message from websocket connection", err)
			break
		}

		msgToBroadcast := RoomMessage{
			author: c,
			room:   room,
			Email:  msgIncoming.Email,
			Event:  msgIncoming.Event,
			Data:   msgIncoming.Data,
		}

		shouldUpdateCanvas := imissphp.InArray(msgIncoming.Event, []string{
			"add-piece",
			"update-piece",
			"remove-piece",
			"update-canvas-data",
		})

		// TODO: Should this be protected by race condition?
		if shouldUpdateCanvas {
			canvasData := canvas_model.CanvasData{}
			err := json.Unmarshal(canvas.CanvasData, &canvasData)
			if err != nil {
				logging.LogError("WebSocket", "Reader -- unmarshalling canvas data", err)
			}

			if msgIncoming.Event == "add-piece" {
				logging.LogInfo("WebSocket", "add-piece", nil)
				bytes, err := json.Marshal(msgIncoming.Data)
				if err != nil {
					logging.LogError("WebSocket", "add-piece -- Could not marhsal piece data", err)
					break
				}

				var piece canvas_model.PieceData
				err = json.Unmarshal(bytes, &piece)
				if err != nil {
					logging.LogError("WebSocket", "add-piece -- Piece data is invalid", err)
					break
				}
				canvasData.PiecesManager.Pieces = append(canvasData.PiecesManager.Pieces, &piece)

			} else if msgIncoming.Event == "update-piece" {
				logging.LogInfo("WebSocket", "update-piece", nil)
				bytes, err := json.Marshal(msgIncoming.Data)
				if err != nil {
					logging.LogError("WebSocket", "add-piece -- Could not marhsal piece data", err)
					break
				}

				var piece canvas_model.PieceData
				err = json.Unmarshal(bytes, &piece)
				if err != nil {
					logging.LogError("WebSocket", "update-piece -- Piece data is invalid", err)
					break
				}

				if indexFloat, ok := (msgIncoming.Data)["index"].(float64); ok {
					canvasData.PiecesManager.Pieces[int(indexFloat)] = &piece
				} else {
					logging.LogError("WebSocket", "msg", msgIncoming.Data)
					break
				}

			} else if msgIncoming.Event == "remove-piece" {
				logging.LogInfo("WebSocket", "remove-piece", nil)
				if indexFloat, ok := (msgIncoming.Data)["index"].(float64); ok {
					i := int(indexFloat)
					if i >= 0 && i < len(canvasData.PiecesManager.Pieces) {
						canvasData.PiecesManager.Pieces = slices.Delete(canvasData.PiecesManager.Pieces, i, i+1)
					}
				} else {
					logging.LogError("WebSocket", "msg", msgIncoming.Data)
					break
				}

			} else if msgIncoming.Event == "update-canvas-data" {
				logging.LogInfo("WebSocket", "update-canvas-data", nil)
				bytes, err := json.Marshal(msgIncoming.Data["canvas_data"])
				if err != nil {
					logging.LogError("WebSocket", "add-piece -- Could not marhsal piece data", err)
					break
				}

				var incomingCanvasData canvas_model.CanvasData
				err = json.Unmarshal(bytes, &incomingCanvasData)
				if err != nil {
					logging.LogError("WebSocket", "add-piece -- Piece data is invalid", err)
					break
				}

				canvasData.BackgroundColor = incomingCanvasData.BackgroundColor

				if c.userUuid == canvas.UserUuid {
					// Only canvas owner allowed:
					canvasData.Name = incomingCanvasData.Name
				}

				canvasDataMapStringAny := service.ToMapStringAny(canvasData)
				msgToBroadcast.Data = map[string]any{
					"canvas_data": canvasDataMapStringAny,
				}
			}

			canvasDataBytes, err := json.Marshal(canvasData) // TODO: make Canvas.CanvasData the actual CanvasData type?
			if err != nil {
				// TODO: Error messages of this and above
				logging.LogError("WebSocket", "Error marshalling canvas data", err)
				break
			}
			canvas.CanvasData = canvasDataBytes

			// If needed, save updates to the canvas
			tx, err := database_config.DB(ctx)
			defer tx.Commit()
			if err != nil {
				tx.Rollback()
				break
			}
			err = canvas.Save(tx)
			if err != nil {
				logging.LogError("WebSocket", "Error saving canvas data after receiving message", err)
				tx.Rollback()
				break
			}
			tx.Commit() // Commit the transaction
		}

		// Fwd msg to other clients in the room
		Broadcast(msgToBroadcast)
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
