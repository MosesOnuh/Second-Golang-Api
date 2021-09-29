package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"log"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/satori/go.uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
)

const (
	DbName         = "usersdb"
	TaskCollection = "tasks"
	UserCollection = "users"

	jwtSecret = "secretname"
)

type User struct {
	ID       string    `json:"id" bson:"id"`
	Name     string    `json:"name" bson:"name"`
	Email    string     `json: "email" bson: "email"`
	Password string    `json:"password" bson:"password"`
	Ts       time.Time `json:"timestamp" bson:"timestamp"`
}

type Task struct {
	ID          string    `json:"id"`
	Owner       string    `json:"owner"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Ts          time.Time `json: "timestamp"`
}

type Claims struct {
	UserId string `json:"user_id"`
	jwt.StandardClaims
}

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

		log.Fatalf("Mongo db not available: %v\n", err)
	}

	router := gin.Default()
	router.GET("/", welcomeHandler)
	router.POST("/createTask", createTaskHandler)
	router.GET("/getTask/:id", getSingleTaskHandler)
	router.GET("/getTasks", getAllTasksHandler)
	router.PATCH("/updateTask/id", updateTaskHandler)
	router.DELETE("/deleteTask/:name", deleteTaskHandler)
	router.POST("/login", loginHandler)
	router.POST("/signup", signupHandler)

	

	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	_ = router.Run(":" + port)

}

func welcomeHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "welcome to task manager API",
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


func createTaskHandler(c *gin.Context) {
	authorization := c.Request.Header.Get("Authorization")

	if authorization == "" {
		c.JSON(401, gin.H{
			"error": "auth token not supplied",
		})
		return
	}

	jwtToken := ""
	splitTokenArray := strings.Split(authorization, "")
	if len(splitTokenArray) > 1 {
		jwtToken = splitTokenArray[1]
	}

	claims := &Claims{}

	keyFunc := func(token *jwt.Token) (i interface{}, e error){
		return []byte(jwtSecret), nil
	}
	token, err := jwt.ParseWithClaims(jwtToken, claims, keyFunc)

	if !token.Valid {
		c.JSON(400, gin.H{
			"error": "invalid jwt token",
		})
		return
	}




	var taskReq Task

	err = c.ShouldBindJSON(&taskReq)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "invalid request data",
		})
		return
	}

	taskId := uuid.NewV4().String()

	task := Task{
		ID: taskId,
		Owner: claims.UserId,
		Name: taskReq.Name,
		Description: taskReq.Description,
		Ts: time.Now(),
	}

	_, err = dbClient.Database("DbName").Collection(TaskCollection).InsertOne(context.Background(), task)
	if err != nil {
		fmt.Println("error saving task", err)
		c.JSON(500, gin.H{
			"error": "Could not process request, could not save task",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "successfully created task",
		"data":    task,
	})
}





func getSingleTaskHandler(c *gin.Context) {
	taskId := c.Param("id")

	var task Task
	query := bson.M{
		"id": taskId,
	}
	err := dbClient.Database("DbName").Collection("TaskCollection").FindOne(context.Background(), query).Decode(&task)

	if err != nil {
		fmt.Println("task not found", err)
		c.JSON(404, gin.H{
			"error": "invalid task id: " + taskId,
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "success",
		"data":    task,
	})
}

func getAllUserHandler(c *gin.Context) {
	var users []User

	cursor, err := dbClient.Database("FirstDB").Collection("Users").Find(context.Background(), bson.M{})
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Could not process request, could not get users",
		})
		return
	}
	err = cursor.All(context.Background(), &users)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Could not process request, could not get users",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data":    users,
	})
}
func getAllTasksHandler(c *gin.Context) {
	authorization := c.Request.Header.Get("Authorization")
	if authorization == ""{
		c.JSON(401, gin.H{
			"error": "auth token required",
		})
		return
	}
	
	jwtToken := ""
	sp := strings.Split(authorization, " ")
	if len(sp) > 1 {
		jwtToken = sp[1]
	}

	// decode token to get claims
	claims := &Claims{}
	keyFunc := func(token *jwt.Token) (i interface{}, e error) {
		return []byte(jwtSecret), nil
	}

	token, err := jwt.ParseWithClaims(jwtToken, claims, keyFunc)
	if !token.Valid {
		c.JSON(401, gin.H{
			"error": "invalid jwt token",
		})
		return
	}
	
	var tasks []Task
	query := bson.M{
		"owner": claims.UserId,
	}

	cursor, err := dbClient.Database("DbName").Collection("TaskCollection").Find(context.Background(), query)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Could not process request, could not get tasks",
		})
		return
	}
	err = cursor.All(context.Background(), &tasks)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Could not process request, could not get users",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data":    tasks,
	})
}

func updateTaskHandler(c *gin.Context) {
	taskId := c.Param("id")

	var task Task

	err := c.ShouldBindJSON(&task)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "invalid request data",
		})
		return
	}
	filterQuery := bson.M{
		"id": taskId,
	}

	updateQuery := bson.M{
		"$set": bson.M{
			"name":  task.Name,
			"description":   task.Description,
		},
	}

	_, err = dbClient.Database(DbName).Collection(TaskCollection).UpdateOne(context.Background(), filterQuery, updateQuery)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Could not process request, could not update task",
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "Task updated!",
	})
}

func deleteTaskHandler(c *gin.Context) {
	taskId := c.Param("id")
	query := bson.M{
		"id": taskId,
	}
	_, err := dbClient.Database("DbName").Collection("TaskCollection").DeleteOne(context.Background(), query)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "Could not process request, could not delete task",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "Task deleted!",
	})
}


func signupHandler(c *gin.Context) {
	type SignupRequest struct {
		Name     string `json:name`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var signupReq SignupRequest
	err := c.ShouldBindJSON(&signupReq)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "invalid request data",
		})
		return
	}
	query := bson.M{
		"email": signupReq.Email,
	}

	count, err := dbClient.Database(DbName).Collection(UserCollection).CountDocuments(context.Background(), query)
	if err != nil {
		fmt.Println("error searching for user: ", err)
		c.JSON(500, gin.H{
			"error": "Could not process reques, please try again later",
		})
		return
	}
	if count > 0 {
		c.JSON(500, gin.H{
			"error": "Email already exits, please use a different email",
		})
		return
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(signupReq.Password), bcrypt.DefaultCost)
	hashPassword := string(bytes)

	userId := uuid.NewV4().String()

	user := User{
		ID:       userId,
		Name:     signupReq.Name,
		Email:    signupReq.Email,
		Password: hashPassword,
		Ts:       time.Now(),
	}

	_, err = dbClient.Database(DbName).Collection(UserCollection).InsertOne(context.Background(), user)
if err != nil {
	fmt.Println("error saving user", err)
	c.JSON(500, gin.H{
		"error":"Could not process request, could not save user",
	})
	return
}
claims := &Claims{
	UserId: user.ID,
	StandardClaims: jwt.StandardClaims{
		IssuedAt: time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
	},
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
jwtTokenString, err := token.SignedString([]byte(jwtSecret))

c.JSON(200, gin.H{
	"message": "sign up successful",
	"token": jwtTokenString,
	"data": user,
})


}

func loginHandler(c *gin.Context) {
	loginReq := struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}{}

	err := c.ShouldBindJSON(&loginReq)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "invalid request data",
		})
		return
	}

	var user User
	query := bson.M{
		"email": loginReq.Email,
	}
	err = dbClient.Database(DbName).Collection(UserCollection).FindOne(context.Background(), query).Decode(&user)
	if err != nil {
		fmt.Printf("error getting user from db: %v\n", err)
		c.JSON(500, gin.H{
			"error": "Could not process request, could not get user",
		})
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password))
	if err != nil{
		fmt.Printf("error validating password: %v\n", err)
		c.JSON(500, gin.H{
			"error": "Invalid login details",
		})
		return
	}

	claims := &Claims{
	UserId: user.ID,
	StandardClaims: jwt.StandardClaims{
		IssuedAt: time.Now().Unix(),
		ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
	},
}

token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
jwtTokenString, err := token.SignedString([]byte(jwtSecret))

c.JSON(200, gin.H{
	"message": "login up successful",
	"token": jwtTokenString,
	"data": user,
})


}
