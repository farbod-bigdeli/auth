package auth

import (
	"net/http"
	"strconv"
	"strings"

	"math/rand"

	"example.com/models/database"
	"example.com/models/jwt"
	"example.com/models/models"
	"example.com/models/notification"
	"github.com/gin-gonic/gin"
)


type SignUpPhone struct {
	Phone string `json:"phone" form:"phone" binding:"required"`
}

type VerificationCodes struct {
	VerifyCode	int	`json:"verifyCode"`
	TempCode	int	`json:"tempCode"`
	Type		SMSType
	UserID		int
}

type SMSType string

const(
	Login_sms 	SMSType = "login"
	Signup_sms	SMSType = "signup"
)

type SignUpInputs struct {
	VerificationCodes
	SignUpPhone
	Username  string `json:"username" binding:"required"`
	Ssn       string `json:"ssn" binding:"required"`
	Locale    string `json:"locale" binding:"required"`
	Birthdate string `json:"birthdate" binding:"required"`
}

type Credentials struct {
	Username  	string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
}

type LoginVerifyInputs struct {
	Username		string 	`json:"username" binding:"required"`
	VerifyCode		int		`json:"verifyCode" binding:"required"`
	TempCode		int		`json:"tempCode" binding:"required"`
}

var PhoneVerifications = make(map[string]VerificationCodes)

func SendSignUpSMS(ct *gin.Context) {
	var phoneInput SignUpPhone
	err := ct.ShouldBind(&phoneInput)
	if err != nil {
		ct.JSON(http.StatusBadRequest, gin.H{
			"message": "Wrong Input",
		})
		return
	}
	//TODO sms
	v := VerificationCodes{
		VerifyCode: rand.Intn(99999-10000)+10000,
		TempCode: rand.Intn(99999-10000)+10000,
		Type: Signup_sms,

	}
	PhoneVerifications[phoneInput.Phone] = v
	message := "code: " + strconv.Itoa(v.VerifyCode)
	notification.SendSMS(message,[]string{phoneInput.Phone})
	ct.JSON(http.StatusOK, gin.H{
		"tempCode": v.TempCode,
	})
}

func Signup(ct *gin.Context) {
	var v SignUpInputs
	
	if err := ct.ShouldBind(&v) ; err != nil {
		ct.JSON(http.StatusBadRequest, gin.H{
			"message": "Wrong Input",
		})
		return
	}
	if checkCode(v, Signup_sms){
		//TODO password
		password := strconv.Itoa(rand.Intn(999999-100000)+100000)
		userInfo := &model.UserInfo{
			User: model.User{
				Username: v.Username,
				Password: password,
			},
			Phone: v.Phone,
		}
		db := database.GetExec()

		if db.Create(userInfo).Error == nil {
			token, _ :=jwt.CreateToken(userInfo.UserID);
			ct.JSON(http.StatusOK, gin.H{
				"token": token,
			})
			message := "password: " + password
			notification.SendSMS(message,[]string{v.Phone})
			delete(PhoneVerifications, v.Phone)
			return
		}
	}
	
	ct.JSON(http.StatusBadRequest, gin.H{
		"message": "Failed",
	})


}


func GetUser(ct *gin.Context) {
	if tokenString := extractToken(ct); tokenString != ""{
		if claims, err := jwt.VerifyToken(tokenString); err == nil {
			ct.JSON(http.StatusOK, gin.H{
				"user": claims["user_id"],
			})
			return
		} 

	}

	ct.JSON(http.StatusUnauthorized,"Unauthorized.")
}

func LogOut(ct *gin.Context) {
	if tokenString := extractToken(ct); tokenString != ""{
		jwt.InvalidateToken(tokenString)
		ct.JSON(http.StatusOK, "Successful")
		return

	}
	ct.JSON(http.StatusBadRequest,"")
}

func checkCode(v interface{}, st SMSType) bool {
	switch v := v.(type) {
	case SignUpInputs:
		if val, ok := PhoneVerifications[v.Phone]; ok && val.TempCode == v.TempCode && val.VerifyCode == v.VerifyCode && val.Type == st{
			return true
		}
	case LoginVerifyInputs:
		if val, ok := PhoneVerifications[v.Username]; ok && val.TempCode == v.TempCode && val.VerifyCode == v.VerifyCode && val.Type == st{
			return true
		}
	default: 
		return false
	}


	return false
}



func Login(ct *gin.Context) {

	var credentials Credentials
	err := ct.ShouldBind(&credentials)
	if err != nil {
		ct.JSON(http.StatusBadRequest, gin.H{
			"message": "Wrong Input",
		})
		return
	}
	db := database.GetExec();
	user := &model.User{}
	result := db.Where("username = ? ",credentials.Username).First(&user)
	if result.Error == nil && result.RowsAffected == 1 && user.CheckPassword(credentials.Password) == nil  {

		//TODO sms

		v := VerificationCodes{
			VerifyCode: rand.Intn(99999-10000)+10000,
			TempCode: rand.Intn(99999-10000)+10000,
			Type: Login_sms,
			UserID: int(user.ID),
	
		}
		PhoneVerifications[credentials.Username] = v

		var userInfo model.UserInfo

		result = db.Where("user_id = ? ", user.ID).First(&userInfo)
		if result.Error == nil && result.RowsAffected == 1   {
			ct.JSON(http.StatusOK, gin.H{
				"tempCode": v.TempCode,
			})
			message := "code: " + strconv.Itoa(v.VerifyCode)
			notification.SendSMS(message,[]string{userInfo.Phone})
	
			return
		}

	

	}

	ct.JSON(http.StatusNotFound, "username or password wrong")


}

func VerifyLogin(ct *gin.Context) {
	var v LoginVerifyInputs
	
	if err := ct.ShouldBind(&v) ; err != nil {
		ct.JSON(http.StatusBadRequest, gin.H{
			"message": "Wrong Input",
		})
		return
	}
	if checkCode(v, Login_sms){
		//TODO password

		token, _ := jwt.CreateToken(PhoneVerifications[v.Username].UserID);
		ct.JSON(http.StatusOK, gin.H{
			"token": token,
		})
		delete(PhoneVerifications, v.Username)
		return

	}
	
	ct.JSON(http.StatusBadRequest, gin.H{
		"message": "Failed",
	})

}

func extractToken (ct *gin.Context) string {
	bearToken := ct.GetHeader("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
	   return strArr[1]
	}
	return ""
}