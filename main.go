package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"log"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
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

var dbClient *mongo.Client

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {

		log.Fatalf("Could not connect to the db: %v\n", err)
	}

	dbClient = client
	err = dbClient.Ping(ctx, readpref.Primary())
	if err != nil {

		log.Fatalf("MOngo db not available: %v\n", err)
	}

	router := gin.Default()
	router.GET("/", helloWorldHandler)
	router.POST("/createUser", createUserHandler)
	router.GET("/getUser/:name", getSingleUserHandler)
	router.GET("/getUsers", getAllUserHandler)

	router.PATCH("/updateUser/:name", updateUserHandler)
	router.DELETE("/deleteUser/:name", deleteUserHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	_ = router.Run(":" + port)

}

func helloWorldHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "hello world",
	})
}

func createUserHandler(c *gin.Context) {
	var user User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "invalid request data",
		})
		return
	}
	_, err = dbClient.Database("FirstDB").Collection("Users").InsertOne(context.Background(), user)
	if err != nil {
		fmt.Println("error saving user", err)
		c.JSON(500, gin.H{
			"error": "Could not process request, could not save user",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "successfully created user",
		"data":    user,
	})
}


func getSingleUserHandler(c *gin.Context) {
	name := c.Param("name")

	var user User
	query := bson.M{
		"name": name,
	}
	err := dbClient.Database("FirstDB").Collection("Users").FindOne(context.Background(), query).Decode(&user)

	if err != nil {
		fmt.Println("user not found", err)
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
	var users []User

	cursor, err := dbClient.Database("FirstDB").Collection("Users").Find(context.Background(), bson.M{} )
	if err !=nil {
		c.JSON(500, gin.H{
			"error": "Could not process request, could get users",
		})
		return	
	}
	err = cursor.All(context.Background(), &users)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Could not process request, could get users",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data": users,
	})
}

func updateUserHandler(c *gin.Context) {
	name :=c.Param("name")

	var user User

	err := c.ShouldBindJSON(&user)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "invalid request data",
		})
		return
	}
	filterQuery := bson.M{
		"name": name,
	}

	updateQuery := bson.M{
		"$set": bson.M{
			"name": user.Name,
			"age": user.Age,
			"email": user.Email,
		},
	}

		_, err = dbClient.Database("FirstDB").Collection("Users").UpdateOne(context.Background(), filterQuery, updateQuery)
			if err != nil {
				c.JSON(500, gin.H{
					"error": "Could not process request, could not update user",
				})
				return
			}

			c.JSON(200, gin.H{
				"message": "User updated!",
			})
 }

 func deleteUserHandler (c *gin.Context) {
	 name := c.Param("name")
	 query := bson.M{
		 "name": name,
	 }
	 _, err := dbClient.Database("FirstDB").Collection("Users").DeleteOne(context.Background(), query)

	 	if err != nil {
			 c.JSON(500, gin.H{
				 "error": "Could not process request, could not delete user",
			 })
			 return
		 }
		 c.JSON(200, gin.H{
			 "message": "user deleted",
		 })
 }
