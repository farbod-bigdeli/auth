package main

import (
	"example.com/models/auth"
	"example.com/models/database"
	"example.com/models/proof"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

)

func main() {
	
	router := gin.Default()
	router.POST("auth/sendSMS", auth.SendSignUpSMS)
	router.POST("auth/signup", auth.Signup)
	router.POST("auth/login", auth.Login)
	router.POST("auth/login_verify", auth.VerifyLogin)
	router.POST("auth/proof/store", proof.Store)

	router.GET("auth/validate", auth.GetUser)
	router.GET("auth/user/get", auth.GetUserWithToken)

	router.GET("auth/logout", auth.LogOut)
	//TODO remove init
	router.GET("auth/init", func(c *gin.Context) {
		database.InitDB()
	})


	router.Run(":8002")
}


func init() {
	if err := godotenv.Load(".env"); err != nil {
		panic(err)
	}
}

