package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/robithritz/chirpbird/chats"
	"github.com/robithritz/chirpbird/common/middleware"
	"github.com/robithritz/chirpbird/common/websocket"
	"github.com/robithritz/chirpbird/users"
)

type SimpleMessage struct {
	Status  bool         `json:"status"`
	Message string       `json:"message"`
	Data    []users.User `json:"data"`
}

type SimpleSingleMessage struct {
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    *users.User `json:"data"`
}

type Login struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func StartServer() {
	go websocket.H.Run()

	router := gin.Default()
	router.LoadHTMLGlob("templates/*.html")
	router.Static("/assets", "./templates")

	router.GET("/", homePage)
	router.GET("/login", loginPage)
	router.GET("/register", registerPage)
	router.GET("/ws", wsConnect)

	router.POST("/users", createUser)
	router.POST("/login", login)

	authorized := router.Group("/")

	authorized.Use(middleware.AuthorizeJWT())
	{
		authorized.GET("check-token", middleware.CheckToken)
		usersAuthorized := authorized.Group("users")
		usersAuthorized.GET("", getUsers)
		usersAuthorized.GET(":id", getSingleUser)

		chatsAuthorized := authorized.Group("chats")
		chatsAuthorized.POST("room", createRoom)
		chatsAuthorized.GET("room/:room_id", getRoomInfo)

	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run("localhost:" + port)
}

func homePage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index.html", nil)
}
func loginPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "login.html", nil)
}
func registerPage(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "register.html", nil)
}

func wsConnect(ctx *gin.Context) {
	websocket.ServeWs(ctx.Writer, ctx.Request)
}

func createUser(ctx *gin.Context) {
	var obj users.UserCreate
	decoder := json.NewDecoder(ctx.Request.Body)

	err := decoder.Decode(&obj)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadGateway, gin.H{
			"status":  false,
			"message": "something went wrong, " + err.Error(),
		})
		return
	}
	if obj.Name == "" || obj.Username == "" || obj.Password == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "field incomplete",
		})
		return
	}

	newId, errors := users.AddNewUser(obj)
	if errors != nil {
		ctx.IndentedJSON(http.StatusBadGateway, gin.H{
			"status":  false,
			"message": "something went wrong, " + errors.Error(),
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "User successfuly created",
		"id":      newId,
	})
}

func getUsers(ctx *gin.Context) {

	search := ctx.Query("s")
	var obj SimpleMessage

	result, err := users.SearchUsers(search)
	if err != nil {
		fmt.Println(err)
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	obj.Status = true
	obj.Message = "Successful"
	obj.Data = result
	ctx.IndentedJSON(http.StatusOK, obj)
}

func getSingleUser(ctx *gin.Context) {
	id := ctx.Param("id")
	var obj SimpleSingleMessage

	if id == "" {
		obj.Status = false
		obj.Message = "field id required"
		ctx.IndentedJSON(http.StatusBadRequest, obj)
		return
	}

	convId, err := strconv.Atoi(id)
	if err != nil {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	result, err := users.GetUser(convId)
	if err != nil {
		fmt.Println(err)
		obj.Status = false
		obj.Message = err.Error()
		if strings.Contains(err.Error(), "no rows") {
			ctx.IndentedJSON(http.StatusNotFound, obj)
		} else {
			ctx.IndentedJSON(http.StatusBadGateway, obj)
		}
		return
	}

	obj.Status = true
	obj.Message = "Successful"
	obj.Data = &result
	ctx.IndentedJSON(http.StatusOK, obj)
}

func login(ctx *gin.Context) {
	var data Login
	decoder := json.NewDecoder(ctx.Request.Body)

	err := decoder.Decode(&data)
	if err != nil {
		fmt.Println(err.Error())
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": err.Error(),
		})
	}

	result, err := users.Authenticate(data.Username, data.Password)
	if err != nil {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	resp := gin.H{
		"status":  true,
		"message": "Successful",
		"token":   result,
	}

	ctx.IndentedJSON(http.StatusOK, resp)
}

func createRoom(ctx *gin.Context) {
	var room chats.Room
	decoder := json.NewDecoder(ctx.Request.Body)

	err := decoder.Decode(&room)
	if err != nil {
		ctx.IndentedJSON(http.StatusBadGateway, gin.H{
			"status":  false,
			"message": "something went wrong, " + err.Error(),
		})
		return
	}
	if room.RoomType == "" || len(room.Participants) == 0 {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "field incomplete",
		})
		return
	}

	ctxInterface, exist := ctx.Get("username")
	if !exist {
		fmt.Println(err)
	}
	username := fmt.Sprintf("%v", ctxInterface)
	room.CreatedBy = username

	newId, errors := chats.CreateRoom(room)
	if errors != nil {
		ctx.IndentedJSON(http.StatusBadGateway, gin.H{
			"status":  false,
			"message": "something went wrong, " + errors.Error(),
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Room successfuly created",
		"room_id": newId,
	})
}

func getRoomInfo(ctx *gin.Context) {
	roomId := ctx.Param("room_id")

	if roomId == "" {
		ctx.IndentedJSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "field room_id required",
		})
		return
	}

	roomIdAsInt, err := strconv.Atoi(roomId)
	if err != nil {
		ctx.IndentedJSON(http.StatusUnauthorized, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	result, err := chats.GetRoomInfo(roomIdAsInt)
	if err != nil {
		fmt.Println(err)
		ctx.IndentedJSON(http.StatusBadGateway, gin.H{
			"status":  false,
			"message": err.Error(),
		})
		return
	}

	ctx.IndentedJSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Successful",
		"data":    result,
	})
}
