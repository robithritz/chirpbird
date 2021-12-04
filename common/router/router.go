package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/robithritz/chirpbird/common/middleware"
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
	router := gin.Default()
	router.GET("/", handler)

	authorized := router.Group("/")

	authorized.Use(middleware.AuthorizeJWT())
	{
		usersAuthorized := authorized.Group("users")
		usersAuthorized.GET("", getUsers)
		usersAuthorized.GET(":id", getSingleUser)
	}

	router.POST("/users", createUser)
	router.POST("/login", login)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	router.Run("localhost:" + port)
}

func handler(ctx *gin.Context) {
	ctx.IndentedJSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Welcome, Chirpbird API v1",
	})
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
		obj.Status = false
		obj.Message = err.Error()

		ctx.IndentedJSON(http.StatusBadGateway, obj)
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
		obj.Status = false
		obj.Message = err.Error()

		fmt.Println(err.Error())
		ctx.IndentedJSON(http.StatusBadRequest, obj)
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
		resp := SimpleMessage{
			Status:  false,
			Message: err.Error(),
		}

		ctx.IndentedJSON(http.StatusBadGateway, resp)
	}

	result, err := users.Authenticate(data.Username, data.Password)
	if err != nil {
		resp := gin.H{
			"status":  false,
			"message": err.Error(),
		}

		ctx.IndentedJSON(http.StatusUnauthorized, resp)
		return
	}

	resp := gin.H{
		"status":  true,
		"message": "Successful",
		"token":   result,
	}

	ctx.IndentedJSON(http.StatusOK, resp)

}
