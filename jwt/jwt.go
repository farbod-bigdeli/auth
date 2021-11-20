package jwt

import (
	"fmt"
	"time"
	"os"

	"github.com/dgrijalva/jwt-go"
)

var tokens = map[string]bool{
	"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdXRob3JpemVkIjp0cnVlLCJleHAiOjI4MzYyNzc3NzAsInVzZXJfaWQiOjg2fQ.11XZS_P8Q3o-PhdWY0B8f6nTnrUpPoDl_9OlKVWUlmY": true,
}

func CreateToken(userId int) (string, error) {
	atClaims := jwt.MapClaims{
		"authorized": true,
		"user_id": userId,
		"exp": time.Now().Add(time.Hour * 60).Unix(),
	}
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		panic("jwt key not found")
	}	
	token,err := at.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	tokens[token] = true
	return token, nil 
}




func InvalidateToken(token string) {
	if isValidToken(token){
		tokens[token] = false
	}
}

func isValidToken(token string) bool {
	if val, ok := tokens[token]; ok{
		return val
	}
	return false
}





func VerifyToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
	   if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		  return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
	   }
	   secret := os.Getenv("JWT_SECRET")
	   if secret == "" {
		   panic("key not found")
	   }
	   return []byte(secret), nil
	})
	if err != nil {
	   return nil, err
	}

	if v, ok := token.Claims.(jwt.MapClaims); ok && token.Valid && isValidToken(tokenString) {
		return v, nil
	 }
	return nil, fmt.Errorf("token not valid")
}


