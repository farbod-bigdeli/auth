package main

import (
	"fmt"
	"net/http"

	"example.com/models/auth"
	"example.com/models/config"
	"example.com/models/database"
	"example.com/models/proof"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		panic(err)
	}
	fmt.Println(config.Config)
	if err := godotenv.Load(config.Config.ENVPath); err != nil {
		panic(err)
	}
	database.GetExec()
	router := gin.Default()
	router.POST("auth/sendSMS", auth.SendSignUpSMS)
	router.POST("auth/signup", auth.Signup)
	router.POST("auth/login", auth.Login)
	router.POST("auth/login_verify", auth.VerifyLogin)
	router.POST("auth/forgot_password", auth.SendForgotPasswordSMS)
	router.POST("auth/verify_forgot_password", auth.VerifyForgotPassword)
	needAuth := router.Group("/")
	needAuth.Use(authorize())
	{
		needAuth.POST("auth/proof/store", proof.Store)
		needAuth.GET("auth/validate", auth.ValidateUser)
		needAuth.GET("auth/user/get", auth.GetUserWithToken)
		needAuth.GET("auth/user/bank", proof.GetBankAccount)
		needAuth.POST("auth/change_password", auth.ChangePassword)

		router.GET("auth/logout", auth.LogOut)
	}

	//TODO remove init
	router.GET("auth/init", func(c *gin.Context) {
		
	})
	router.Run(config.Config.LocalPort)
}


func authorize() gin.HandlerFunc {
	return func (ct *gin.Context)  {
		id, err := auth.GetUserId(ct)
		if err != nil {
			ct.AbortWithStatusJSON(http.StatusUnauthorized, "Unathoized.")
			return
		}
		ct.Set("userId", id)
		ct.Next()
	}
}

