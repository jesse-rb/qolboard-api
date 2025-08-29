package websocket_service

import (
	"encoding/json"
	"net/http"
	database_config "qolboard-api/config/database"
	model "qolboard-api/models"
	service "qolboard-api/services"
	auth_service "qolboard-api/services/auth"
	canvas_service "qolboard-api/services/canvas"
	"qolboard-api/services/logging"
	response_service "qolboard-api/services/response"
	"slices"
	"time"

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

type dataChJoin struct {
	userUuid string
	canvas   *model.Canvas
	conn     *websocket.Conn
	chResume chan *Client
}

type RoomsManager struct {
	roomsMap    map[uint64]*Room
	chJoin      chan *dataChJoin
	chLeave     chan *Client
	chBroadcast chan RoomMessage
}

type Room struct {
	Canvas  *model.Canvas
	Clients map[*Client]bool
	chSave  chan bool
	chClose chan bool
}

type Client struct {
	userUuid string
	room     *Room
	conn     *websocket.Conn
	chSend   chan RoomMessage
}

type RoomMessage struct {
	author *Client
	room   *Room
	Event  string         `json:"event" binding:"required"`
	Email  string         `json:"email" binding:"required"`
	Data   map[string]any `json:"data" binding:"required"`
}

var rm *RoomsManager = NewRoomsManager()

func init() {
	go rm.Run()
}

func NewRoomsManager() *RoomsManager {
	return &RoomsManager{
		roomsMap:    make(map[uint64]*Room),
		chJoin:      make(chan *dataChJoin),
		chLeave:     make(chan *Client),
		chBroadcast: make(chan RoomMessage),
	}
}

func NewRoom(canvas *model.Canvas) *Room {
	return &Room{
		Canvas:  canvas,
		Clients: make(map[*Client]bool),
		chSave:  make(chan bool),
		chClose: make(chan bool),
	}
}

func NewClient(userUuid string, room *Room, conn *websocket.Conn) *Client {
	client := &Client{
		userUuid: userUuid,
		room:     room,
		chSend:   make(chan RoomMessage, 256), // Allow for some buffer
		conn:     conn,
	}

	return client
}

func BroadcastCanvas(canvas *model.Canvas) {
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
	go room.Run()

	return room
}

func (rm *RoomsManager) Run() {
	// Event loop for our rooms manager, only one of these events should run at a given time
	for {
		select {
		// A client joins a room
		case joinRoomData := <-rm.chJoin:
			logging.LogDebug("(WS event loop)", "receiving join", nil)
			room := rm.getRoom(joinRoomData.canvas)
			client := NewClient(joinRoomData.userUuid, room, joinRoomData.conn)
			room.addClient(client)
			joinRoomData.chResume <- client // Send the client back to the websocket connection controller action

		// A client leaves a room
		case client := <-rm.chLeave:
			logging.LogDebug("(WS event loop)", "receiving leave", nil)
			room := client.room
			room.removeClient(client)
			close(client.chSend)
			client.conn.Close()

			if !room.hasClients() {
				// If the room is now empty, do some cleanup by deleting the room
				room.chSave <- true
				room.chClose <- true
				delete(rm.roomsMap, room.Canvas.ID)
			}

		case msg := <-rm.chBroadcast:
			for c := range msg.room.Clients {
				if msg.author == c {
					continue // Don't send the message back to the author
				}
				select {
				// Attempt to send message to the cleint (YAY go channels!)
				case c.chSend <- msg:
				// client's send queue is full, better not hold up our entire event loop...
				default:
					logging.LogDebug("(WS event loop)", "skipping... client send channel is FULL", map[string]any{
						"available cap": cap(c.chSend),
						"queued len":    len(c.chSend),
					})
					continue
				}
			}
		}
	}
}

func (room *Room) Run() {
	// Ticker to save canvas every interval
	ticker := time.NewTicker(30 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				room.chSave <- true
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	// Event loop for room
	for {
		select {
		case meh := <-room.chSave:
			logging.LogDebug("WebSocket", "Saving canvas triggered", meh)
			tx, err := database_config.DB(nil)
			defer tx.Rollback()
			if err != nil {
				break
			}
			err = room.Canvas.SystemUpdate(tx)
			if err != nil {
				logging.LogError("WebSocket", "Error saving canvas data after receiving message", err)
				break
			}
			tx.Commit() // Commit the transaction
			logging.LogInfo("WebSocket", "Canvas saved by websocket room", nil)

		case <-room.chClose:
			return
		}
	}
}

// Allow broadcasting to a room externally
func Broadcast(msg RoomMessage) {
	// Broadcast message to all clients in the room
	rm.chBroadcast <- msg
}

func Join(userUuid string, canvas *model.Canvas, conn *websocket.Conn, chResume chan *Client) {
	rm.chJoin <- &dataChJoin{
		userUuid: userUuid,
		canvas:   canvas,
		conn:     conn,
		chResume: chResume,
	}
}

func (c *Client) Leave() {
	rm.chLeave <- c
}

func (room *Room) updateCanvas(ctx *gin.Context, msgIncoming RoomMessage) {
	canvasData := room.Canvas.CanvasData

	if msgIncoming.Event == "add-piece" {
		bytes, err := json.Marshal(msgIncoming.Data)
		if err != nil {
			logging.LogError("WebSocket", "add-piece -- Could not marhsal piece data", err)
			return
		}

		var piece canvas_service.PieceData
		err = json.Unmarshal(bytes, &piece)
		if err != nil {
			logging.LogError("WebSocket", "add-piece -- Piece data is invalid", err)
			return
		}
		canvasData.PiecesManager.Pieces = append(canvasData.PiecesManager.Pieces, &piece)

	} else if msgIncoming.Event == "update-piece" {
		bytes, err := json.Marshal(msgIncoming.Data)
		if err != nil {
			logging.LogError("WebSocket", "update-piece -- Could not marhsal piece data", err)
			return
		}

		var piece canvas_service.PieceData
		err = json.Unmarshal(bytes, &piece)
		if err != nil {
			logging.LogError("WebSocket", "update-piece -- Piece data is invalid", err)
			return
		}

		if indexFloat, ok := (msgIncoming.Data)["index"].(float64); ok {
			index := int(indexFloat)
			if index >= 0 && index < len(canvasData.PiecesManager.Pieces) {
				canvasData.PiecesManager.Pieces[int(indexFloat)] = &piece
			}
		} else {
			logging.LogError("WebSocket", "update-piece -- Failed to update piece", msgIncoming.Data)
			return
		}

	} else if msgIncoming.Event == "remove-piece" {
		if indexFloat, ok := (msgIncoming.Data)["index"].(float64); ok {
			index := int(indexFloat)
			if index >= 0 && index < len(canvasData.PiecesManager.Pieces) {
				canvasData.PiecesManager.Pieces = slices.Delete(canvasData.PiecesManager.Pieces, index, index+1)
			}
		} else {
			logging.LogError("WebSocket", "remove-piece -- Failed to remove piece", msgIncoming.Data)
			return
		}
	} else if msgIncoming.Event == "update-canvas-data" {
		bytes, err := json.Marshal(msgIncoming.Data["canvas_data"])
		if err != nil {
			logging.LogError("WebSocket", "update-canvas-data -- Could not marhsal piece data", err)
			return
		}

		var incomingCanvasData canvas_service.CanvasData
		err = json.Unmarshal(bytes, &incomingCanvasData)
		if err != nil {
			logging.LogError("WebSocket", "update-canvas-data -- Piece data is invalid", err)
			return
		}

		canvasData.BackgroundColor = incomingCanvasData.BackgroundColor

		if auth_service.Auth(ctx) == room.Canvas.UserUuid {
			// Only canvas owner allowed:
			canvasData.Name = incomingCanvasData.Name
		}
	}

	room.Canvas.CanvasData = canvasData
}

func (c *Client) Reader(ctx *gin.Context) {
	logging.LogDebug("WebSocket", "Reader -- Starting reader for client", nil)

	for {
		msgIncoming := RoomMessage{}
		err := c.conn.ReadJSON(&msgIncoming)
		if err != nil {
			logging.LogError("WebSocket", "Error reading message from websocket connection", err)
			break
		}

		msgToBroadcast := RoomMessage{
			author: c,
			room:   c.room,
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

		if shouldUpdateCanvas {
			c.room.updateCanvas(ctx, msgIncoming)

			if msgIncoming.Event == "update-canvas-data" {
				canvasDataMap := service.ToMapStringAny(c.room.Canvas.CanvasData)
				msgToBroadcast.Data = map[string]any{
					"canvas_data": canvasDataMap,
				}
			}
		}

		// Fwd msg to other clients in the room
		Broadcast(msgToBroadcast)
	}
}

func (c *Client) Writer() {
	for msg := range c.chSend {
		err := c.conn.WriteJSON(msg)
		if err != nil {
			logging.LogError("WebSocket", "Error writing message to websocket connection", err)
			continue
		}
	}
}

func (c *Client) Send(msg RoomMessage) {
	c.chSend <- msg
}

func (c *Client) GetRoom() *Room {
	return c.room
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
