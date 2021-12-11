package chats

import (
	"context"
	"fmt"
	"strings"

	"github.com/robithritz/chirpbird/common/database"
)

type Room struct {
	RoomId       int      `json:"room_id"`
	RoomType     string   `json:"room_type"`
	Participants []string `json:"participants"`
	CreatedBy    string   `json:"created_by"`
}

func CreateRoom(data Room) (roomId int, fa error) {

	err := database.DB.QueryRow(context.Background(), "INSERT INTO master_rooms(room_type, participants, created_by) VALUES($1, $2, $3) RETURNING room_id", data.RoomType, strings.Join(data.Participants, ","), data.CreatedBy).Scan(&data.RoomId)
	if err != nil {
		return 0, err
	}

	return data.RoomId, nil
}

func GetListParticipants(roomId int) []string {
	var participants []string
	rows, err := database.DB.Query(context.Background(), "SELECT participants FROM master_rooms WHERE room_id = $1", roomId)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	for rows.Next() {
		var joinedParticipants string

		if err := rows.Scan(&joinedParticipants); err != nil {
			fmt.Println(err)
		}
		participants = strings.Split(joinedParticipants, ",")
	}

	return participants
}

func GetRoomInfo(roomId int) (Room, error) {
	var room Room
	var participantsJoined string

	err := database.DB.QueryRow(context.Background(), "SELECT room_id, room_type, participants, created_by FROM master_rooms WHERE room_id = $1", roomId).Scan(&room.RoomId, &room.RoomType, &participantsJoined, &room.CreatedBy)
	if err != nil {
		fmt.Println(err)
		return room, err
	}
	room.Participants = strings.Split(participantsJoined, ",")

	return room, nil
}
