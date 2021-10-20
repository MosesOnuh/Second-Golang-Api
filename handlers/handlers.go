package handlers

import (
	"fmt"
	"strings"
	"time"

	"github.com/MosesOnuh/todoTask-Api/auth"
	"github.com/MosesOnuh/todoTask-Api/db"
	"github.com/MosesOnuh/todoTask-Api/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	jwtSecret = "secretname"
)

func WelcomeHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "welcome to task manager API",
	})
}

func CreateTaskHandler(c *gin.Context) {
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
	claims, err := auth.ValidToken(jwtToken)
	if err != nil {
		c.JSON(401, gin.H{
			"error": "invalid jwt token",
		})
		return
	}

	var taskReq models.Task

	err = c.ShouldBindJSON(&taskReq)
	if err != nil {
		c.JSON(200, gin.H{
			"error": "Invalid request data",
		})
		return
	}

	taskId := uuid.NewV4().String()

	task := models.Task{
		ID:          taskId,
		Owner:       claims.UserId,
		Name:        taskReq.Name,
		Description: taskReq.Description,
		Ts:          time.Now(),
	}

	_, err = db.CreateTask(&task)
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
func GetSingleTaskHandler(c *gin.Context) {
	taskId := c.Param("id")

	task, err := db.GetSingleTask(taskId)
	if err != nil {
		fmt.Println("user not found", err)
		c.JSON(404, gin.H{
			"error": "invalid task id:" + taskId,
		})
		return
	}

	c.JSON(200, gin.H{
		"message": "success",
		"data":    task,
	})
}

func GetAllUserHandler(c *gin.Context) {
	users, err := db.GetAllUsers()
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Could not process request, users not found",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "success",
		"data":    users,
	})
}

func GetAllTasksHandler(c *gin.Context) {
	authorization := c.Request.Header.Get("Authorization")
	if authorization == "" {
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

	claims := &models.Claims{}
	keyFunc := func(token *jwt.Token) (i interface{}, e error) {
		return []byte(jwtSecret), nil
	}

	token, err := jwt.ParseWithClaims(jwtToken, claims, keyFunc)
	if err != nil {
		c.JSON(401, gin.H{
			"error": "invalid jwt token",
		})
		return
	}
	if !token.Valid {
		c.JSON(401, gin.H{
			"error": "invalid jwt token",
		})
		return
	}


	tasks, err := db.GetAllTasks(claims.UserId)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Could not process request, could not get tasks",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "success",
		"data":    tasks,
	})
}

func UpdateTaskHandler(c *gin.Context) {
	taskId := c.Param("id")

	var task models.Task
	err := c.ShouldBindJSON(&task)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "invalid request data",
		})
		return
	}

	err = db.UpdateTask(taskId, task.Name, task.Description)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "Could not process request, could not update task",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "Task updated",
	})
}
func DeleteTaskHandler(c *gin.Context) {
	authorization := c.Request.Header.Get("Authorization")
	if authorization == " " {
		c.JSON(401, gin.H{
			"error": "auth token supplied",
		})
		return
	}

	jwtToken := ""

	splitTokenArray := strings.Split(authorization, " ")
	if len(splitTokenArray) > 1 {
		jwtToken = splitTokenArray[1]
	}

	claims, err := auth.ValidToken(jwtToken)
	if err != nil {
		c.JSON(401, gin.H{
			"error": "invalid jwt token",
		})
		return
	}

	taskId := c.Param("id")

	err = db.DeleteTask(taskId, claims.UserId)
	if err != nil {
		c.JSON(500, gin.H{
			"error": "could not process, could not delete task",
		})
	}
	c.JSON(200, gin.H{
		"message": "Task deleted",
	})
}
func LoginHandler(c *gin.Context) {
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
	user, err := db.GetUserByEmail(loginReq.Email)
	if err != nil {
		fmt.Printf("error getting user from dn: %v\n", err)
		c.JSON(500, gin.H{
			"error": "Could not process request, could get user",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password))
	if err != nil {
		fmt.Printf("error validating password: %v/n", err)
		c.JSON(500, gin.H{
			"error": "invalid login details",
		})
		return
	}
	jwtTokenString, err := auth.CreateToken(user.ID)

	if err != nil {
		c.JSON(500, gin.H{
			"error": "invalid token",
		})
	}

	c.JSON(200, gin.H{
		"message": "sign up succesful",
		"token":   jwtTokenString,
		"data":    user,
	})

}

func SignupHandler(c *gin.Context) {
	type SignupRequest struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	var signupReq SignupRequest

	err := c.ShouldBindJSON(&signupReq)
	if err != nil {
		c.JSON(400, gin.H{
			"error": "Invalid request data",
		})
		return
	}
	exists := db.CheckUserExists(signupReq.Email)
	if exists {
		c.JSON(500, gin.H{
			"error": "Email already exists, please use a different email",
		})
		return
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(signupReq.Password), bcrypt.DefaultCost)
	if err != nil {
		
		c.JSON(500, gin.H{
			"error": "Could not process request",
		})
		return
	}
	hashPassword := string(bytes)

	userId := uuid.NewV4().String()
	user := models.User{
		ID:       userId,
		Name:     signupReq.Name,
		Email:    signupReq.Email,
		Password: hashPassword,
		Ts:       time.Now(),
	}

	_, err = db.CreateUser(&user)
	if err != nil {
		fmt.Println("error saving user", err)
		c.JSON(500, gin.H{
			"error": "Could not process request could not save user",
		})
		return
	}
	claims := &models.Claims{
		UserId: user.ID,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtTokenString, err := token.SignedString([]byte(jwtSecret))

	if err != nil {
		fmt.Println("error saving user", err)
		c.JSON(500, gin.H{
			"error": "Could not process token",
		})
		return
	}
	c.JSON(200, gin.H{
		"message": "sign up successful",
		"token":   jwtTokenString,
		"data":    "user",
	})

}
