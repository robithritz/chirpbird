package websocket

import (
	"fmt"
	"strconv"

	"github.com/robithritz/chirpbird/chats"
)

type Message struct {
	Data           string
	Room           string
	WriterId       int
	WriterName     string
	WriterUsername string
	CreatedAt      string
}

type subscription struct {
	conn *connection
	// room     string
	id       int
	name     string
	username string
}

type hub struct {
	// rooms map[string]map[*connection]bool

	activeUsers map[string]*connection

	broadcast chan Message

	register chan subscription

	unregister chan subscription
}

var H = hub{
	broadcast:   make(chan Message),
	register:    make(chan subscription),
	unregister:  make(chan subscription),
	activeUsers: make(map[string]*connection),
}

func (h *hub) Run() {
	for {
		select {
		case s := <-h.register:
			h.activeUsers[s.username] = s.conn
		case s := <-h.unregister:
			delete(h.activeUsers, s.username)
		case m := <-h.broadcast:
			roomInt, err := strconv.Atoi(m.Room)
			if err != nil {
				fmt.Println(err)
				return
			}

			listParticipants := chats.GetListParticipants(roomInt)

			for _, participant := range listParticipants {
				if c, ok := h.activeUsers[participant]; ok {
					select {
					case c.send <- m:
					default:
						close(c.send)
						delete(h.activeUsers, participant)
					}
				}
			}

			// connections := h.rooms[m.Room]
			// for c := range connections {
			// 	select {
			// 	case c.send <- m:
			// 	default:
			// 		close(c.send)
			// 		delete(connections, c)
			// 		if len(connections) == 0 {
			// 			delete(h.rooms, m.Room)
			// 		}
			// 	}
			// }
		}
	}
}
