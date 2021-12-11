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
	conn     *connection
	room     string
	id       int
	name     string
	username string
}

// hub maintains the set of active connections and broadcasts messages to the
// connections.
type hub struct {
	// Registered connections.
	rooms map[string]map[*connection]bool

	activeUsers map[string]*connection

	// Inbound messages from the connections.
	broadcast chan Message

	// Register requests from the connections.
	register chan subscription

	// Unregister requests from connections.
	unregister chan subscription
}

var H = hub{
	broadcast:   make(chan Message),
	register:    make(chan subscription),
	unregister:  make(chan subscription),
	rooms:       make(map[string]map[*connection]bool),
	activeUsers: make(map[string]*connection),
}

func (h *hub) Run() {
	for {
		select {
		case s := <-h.register:
			connections := h.rooms[s.room]
			if connections == nil {
				connections = make(map[*connection]bool)
				h.rooms[s.room] = connections
			}
			h.rooms[s.room][s.conn] = true
			h.activeUsers[s.username] = s.conn
			fmt.Println(h.activeUsers)
		case s := <-h.unregister:
			connections := h.rooms[s.room]
			if connections != nil {
				if _, ok := connections[s.conn]; ok {
					delete(connections, s.conn)
					close(s.conn.send)
					if len(connections) == 0 {
						delete(h.rooms, s.room)
					}
				}
			}
			delete(h.activeUsers, s.username)
			fmt.Println(h.activeUsers)
		case m := <-h.broadcast:
			roomInt, err := strconv.Atoi(m.Room)
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println("ada broadcast ke ", roomInt)
			listParticipants := chats.GetListParticipants(roomInt)
			fmt.Println("list parts", listParticipants)
			for i, s := range listParticipants {
				fmt.Println(i, s)
			}
			for _, participant := range listParticipants {
				fmt.Println("part :", participant)
				fmt.Println("conn :", h.activeUsers[participant])
				if c, ok := h.activeUsers[participant]; ok {
					fmt.Println(c)
					select {
					case c.send <- m:
					default:
						fmt.Println("kedelete")
						close(c.send)
						delete(h.activeUsers, participant)
					}
				} else {
					fmt.Println("not oke")
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
