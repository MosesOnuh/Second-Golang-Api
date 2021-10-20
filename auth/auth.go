package auth

import (
	"time"

	"github.com/MosesOnuh/todoTask-Api/models"
	"github.com/dgrijalva/jwt-go"
)

const (
	jwtSecret = "secretname"
)

func ValidToken(jwtToken string) (*models.Claims, error) {
	claims := &models.Claims{}

	keyFunc := func(token *jwt.Token) (i interface{}, e error) {
		return []byte(jwtSecret), nil
	}
	token, err := jwt.ParseWithClaims(jwtToken, claims, keyFunc)

	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, err
	}
	return claims, nil
}

func CreateToken(userId string) (string, error) {
	claims := &models.Claims{
		UserId: userId,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().Unix(),
			ExpiresAt: time.Now().Add(time.Hour * 1).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtTokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", err
	}
	return jwtTokenString, nil
}
