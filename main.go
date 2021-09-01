package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

// multipleUsers := [User{Micheal, James, Timmy}]
// Micheal

type User struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Email   string `json:"email"`
	Country string `json:"country"`
}

var Users []User

func main() {
	router := gin.Default()
	router.GET("/getUser/:name", getSingleHandler)
	router.GET("/getUsers", getAllUserHandler)
	router.POST("/createUsers", createHandler)
	router.PATCH("/updateUsers", updateHandler)
	router.DELETE("/delete", deleteHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	_ = router.Run(":" + port)

}

func getSingleHandler(c *gin.Context) {
	name := c.Param("name")
	fmt.Println("name", name)
	var user User

	userAvailable := false

	for _, value := range Users {

		if value.Name == name {
			user = value
			userAvailable = true
		}
	}

	if !userAvailable {
		c.JSON(404, gin.H{
			"error": "no user with name found: " + name,
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data":    user,
	})
}

func getAllUserHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "success",
		"data":    Users,
	})
}

func createHandler(c *gin.Context) {
	var user User

	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "invalid request data",
		})
		return
	}

	Users = append(Users, user)

	c.JSON(200, gin.H{
		"message": "Profile successfully created",
		"data":    user,
	})
}

func updateHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Profile Updated",
	})
}

func deleteHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "Profile deleted",
	})
}
