package server

import (
	"github.com/MosesOnuh/todoTask-Api/handlers"
	"github.com/gin-gonic/gin"
)

func Run(port string) error {
	router := gin.Default()
	router.GET("/", handlers.WelcomeHandler)
	router.POST("/createTask", handlers.CreateTaskHandler)
	router.GET("/getTask/:id", handlers.GetSingleTaskHandler)
	router.GET("/getTasks", handlers.GetAllTasksHandler)
	router.PATCH("/updateTask/id", handlers.UpdateTaskHandler)
	router.DELETE("/deleteTask/:name", handlers.DeleteTaskHandler)
	router.POST("/login", handlers.LoginHandler)
	router.POST("/signup", handlers.SignupHandler)

	err := router.Run(":" + port)
	if err != nil {
		return err
	}
	return nil
}
