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

// If we needed to scale horizontally one day, this will need some thinking and work
var canvasUserConnMap = make(map[uint64]map[string]*websocket.Conn)

type CanvasMessage struct {
	Event string `json:"event" binding:"required"`
	Email string `json:"email" binding:"required"`
	Data  *gin.H `json:"data" binding:"required"`
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

func AddConnection(canvasId uint64, userUuid string, conn *websocket.Conn) {
	_, ok := canvasUserConnMap[canvasId]
	if !ok {
		canvasUserConnMap[canvasId] = make(map[string]*websocket.Conn, 0)
	}

	canvasUserConnMap[canvasId][userUuid] = conn
}

func WriteToCanvasConnections(canvasId uint64, except *websocket.Conn, data *CanvasMessage) {
	if conns, ok := canvasUserConnMap[canvasId]; ok {
		// Write to each connection
		for _, c := range conns {
			c.WriteJSON(data)
			if c != except {
				c.WriteJSON(data)
			}
		}
	} else {
		logging.LogError("WriteToCanvasConnections", "Attempted to write to canvas connections that is not mapped", canvasId)
	}
}

