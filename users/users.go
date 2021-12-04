package users

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/robithritz/chirpbird/common/database"
	"github.com/robithritz/chirpbird/common/middleware"
	"github.com/robithritz/chirpbird/common/utils"
)

type User struct {
	Id        int    `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type UserCreate struct {
	Username string `json:"username"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func SearchUsers(s string) (result []User, err error) {

	var listUsers []User
	s = "%" + s + "%"
	rows, err := database.DB.Query(context.Background(), "SELECT id, username, name, created_at FROM master_users WHERE username ILIKE $1 OR name ILIKE $1 ORDER BY username ASC LIMIT 10 OFFSET 0", s)
	if err != nil {
		fmt.Println(err)
		return listUsers, err
	}
	defer rows.Close()

	for rows.Next() {
		var obj User
		if err := rows.Scan(&obj.Id, &obj.Username, &obj.Name, &obj.CreatedAt); err != nil {
			fmt.Println(err)
		}
		listUsers = append(listUsers, obj)
	}

	return listUsers, nil
}

func GetUser(id int) (User, error) {
	var user User

	err := database.DB.QueryRow(context.Background(), "SELECT id, username, name, created_at FROM master_users WHERE id = $1", id).Scan(&user.Id, &user.Username, &user.Name, &user.CreatedAt)
	if err != nil {
		fmt.Println(err)
		return user, err
	}

	return user, nil
}

func AddNewUser(data UserCreate) (int, error) {
	var createdId int
	hashedPassword, err := utils.HashPassword(data.Password)
	if err != nil {
		return 0, err
	}

	var usernameExisted bool
	err = database.DB.QueryRow(context.Background(), "SELECT true FROM master_users WHERE username = $1", data.Username).Scan(&usernameExisted)
	if err != nil {

		if strings.Contains(err.Error(), "no rows") {
			goto next
		}
		return 0, err
	}
	if usernameExisted {
		return 0, errors.New("username has already used")
	}

next:

	err = database.DB.QueryRow(context.Background(), "INSERT INTO master_users(username, name, password) VALUES($1, $2, $3) RETURNING id", data.Username, data.Name, hashedPassword).Scan(&createdId)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}

	return createdId, nil
}

func UpdateUser() {
	// reserved
}

func DeleteUser() {
	// reserved
}

func Authenticate(username string, password string) (string, error) {
	var hashedPassword string
	var obj User
	err := database.DB.QueryRow(context.Background(), "SELECT id, username, name, created_at, password FROM master_users WHERE username = $1 AND active IS TRUE", username).Scan(&obj.Id, &obj.Username, &obj.Name, &obj.CreatedAt, &hashedPassword)
	if err != nil {
		fmt.Println(err)
		return "", errors.New("wrong username or password")
	}

	valid, err := utils.VerifyPassword(hashedPassword, password)
	if err != nil || !valid {
		fmt.Println(err)
		return "", errors.New("wrong username or password")
	}

	token, err := middleware.JWTGenToken(obj.Id, obj.Username, obj.Name)
	if err != nil {
		fmt.Println(err)
		return "", errors.New("error token generation")
	}

	return token, nil
}
