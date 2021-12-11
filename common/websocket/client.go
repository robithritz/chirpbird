package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/robithritz/chirpbird/common/middleware"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// connection is an middleman between the websocket connection and the hub.
type connection struct {
	// The websocket connection.
	ws *websocket.Conn

	// Buffered channel of outbound messages.
	send chan Message
}

// readPump pumps messages from the websocket connection to the hub.
func (s subscription) readPump() {
	c := s.conn
	defer func() {
		H.unregister <- s
		c.ws.Close()
	}()
	c.ws.SetReadLimit(maxMessageSize)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error { c.ws.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, msg, err := c.ws.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway) {
				log.Printf("error: %v", err)
			}
			break
		}
		t := time.Now()
		var dat map[string]interface{}
		if err := json.Unmarshal(msg, &dat); err != nil {
			fmt.Println(err)
		}
		message := dat["message"].(string)
		roomId := dat["room_id"].(float64)
		m := Message{message, fmt.Sprintf("%.0f", roomId), s.id, s.name, s.username, t.Format("2006-01-02 15:04:05 -0700")}
		H.broadcast <- m
	}
}

// write writes a message with the given message type and payload.
func (c *connection) writeTicker(mt int, payload []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(mt, payload)
}
func (c *connection) write(mt int, payload Message) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	marshal, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
	}

	return c.ws.WriteMessage(mt, marshal)
}

// writePump pumps messages from the hub to the websocket connection.
func (s *subscription) writePump() {
	c := s.conn
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.ws.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.writeTicker(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.writeTicker(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func ServeWs(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	var token string
	if len(params["token"]) > 0 {
		token = params["token"][0]
	} else {
		log.Println("No Token")
		return
	}

	status, payload, err := middleware.JWTVerifyToken(token)
	if !status || err != nil {
		log.Println("Token not valid")
		return
	}

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err.Error())
		return
	}
	c := &connection{send: make(chan Message, 256), ws: ws}
	s := subscription{c, payload.Id, payload.Name, payload.Username}
	H.register <- s
	// H.rooms[]
	go s.writePump()
	go s.readPump()
}
