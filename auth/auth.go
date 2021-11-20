package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"example.com/models/database"
	"example.com/models/finnotech"

	"example.com/models/jwt"
	"example.com/models/model"
	"example.com/models/notification"
	"example.com/models/util"
	"github.com/gin-gonic/gin"
)


type PhoneInput struct {
	Phone string `json:"phone" form:"phone" binding:"required"`
}

type VerificationCodes struct {
	VerifyCode		int		`json:"verifyCode"`
	TempCode		int		`json:"tempCode"`
	Type			SMSType
	UserID			int
	TrackID			string `json:"trackID"`
	FinnotechToken  string
	Expire			time.Time
}

type SMSType string

const(
	Login_sms 				SMSType = "login"
	Signup_sms				SMSType = "signup"
	Forgot_Password_sms		SMSType = "forgotPassword"
)

type SignUpInputs struct {
	VerificationCodes
	PhoneInput
	Username  string `json:"username" binding:"required"`
	Ssn       string `json:"ssn" binding:"required"`
	Locale    string `json:"locale" binding:"required"`
	Birthdate string `json:"birthdate" binding:"required"`
	Gender	  string `json:"gender" binding:"required"`
	FName	  string `json:"fName" binding:"required"`
	LName	  string `json:"lName" binding:"required"`
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

type ForgotPasswordInput struct {
	Phone			string 	`json:"phone" binding:"required"`
	VerifyCode		int		`json:"verifyCode" binding:"required"`
	TempCode		int		`json:"tempCode" binding:"required"`
}

type ChangePasswordInput struct {
	OldPassword		string 		`json:"oldPassword" binding:"required"`
	NewPassword		string		`json:"newPassword" binding:"required,min=2"`
}

const (
	FinnotechAuthorize 			string = "https://apibeta.finnotech.ir/dev/v2/oauth2/authorize"
	FinnotechVerifySMS 			string = "https://apibeta.finnotech.ir/dev/v2/oauth2/verify/sms"
	FinnotechGetToken 			string = "https://apibeta.finnotech.ir/dev/v2/oauth2/token"
	FinnotechNidVerification 	string = "https://apibeta.finnotech.ir/facility/v2/clients/REDACTED/users/"
)

const (
	CoreNewUser string = "http://YOUR_SERVER:22216/newuser"
)
type FinnotechSendSMSResponse struct {
	Result struct {
		SmsSent bool   `json:"smsSent"`
		TrackID string `json:"trackId"`
	} `json:"result"`
	Status string `json:"status"`
}



var PhoneVerifications = make(map[string]VerificationCodes)

func SendSignUpSMS(ct *gin.Context) {
	var phoneInput PhoneInput
	err := ct.ShouldBind(&phoneInput)
	if err != nil {
		ct.JSON(http.StatusBadRequest, gin.H{
			"message": "Wrong Input",
		})
		return
	}
	//TODO rand
	v := VerificationCodes{
		// VerifyCode: 12345,
		// util.RandInt(10000, 99999),
		TempCode: 12345,
		// util.RandInt(10000, 99999),
		Type: Signup_sms,

	}
	// message := "code: " + strconv.Itoa(v.VerifyCode)
	// notification.SendSMS(message,[]string{phoneInput.Phone})
	// ct.JSON(http.StatusOK, gin.H{
	// 	"tempCode": v.TempCode,
	// })
	client := &http.Client{}
	req, _ := http.NewRequest("GET", FinnotechAuthorize, nil)
	q := req.URL.Query()
	//TODO read env
	q.Add("client_id", "REDACTED")
	q.Add("response_type", "code")
	q.Add("response_type", "code")
	q.Add("redirect_uri", "REDACTED")
	q.Add("scope", "facility:sms-nid-verification:get")
	q.Add("mobile", phoneInput.Phone)
	q.Add("auth_type", "SMS")
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Authorization","bearer " + os.Getenv("FINNOTECH_AUTHORIZE_STRING"))
	res, err := client.Do(req)

	if err != nil {
		return 
	}
	defer res.Body.Close()

	var r FinnotechSendSMSResponse
	json.NewDecoder(res.Body).Decode(&r)

	if res.StatusCode == 200 && r.Status == "DONE" && r.Result.SmsSent {
		v.TrackID = r.Result.TrackID
		fmt.Println(v.TrackID)
		PhoneVerifications[phoneInput.Phone] = v
		ct.JSON(http.StatusOK, gin.H{
			"tempCode": v.TempCode,
		})
		return 
	}
	ct.JSON(http.StatusOK, gin.H{
		"message": "failed",
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
		finnotech := finnotech.NewFinnotech()
		val, ok := PhoneVerifications[v.Phone]
		if !ok || val.FinnotechToken == "" {

			code, err := finnotech.VerifySMS(v.Phone, strconv.Itoa(v.VerifyCode), v.Ssn, PhoneVerifications[v.Phone].TrackID)
	
			if err != nil {
				ct.JSON(http.StatusBadRequest, gin.H{
					"message": "Failed to verify token",
					"status": err.Error(),
				})
				return
			}
			err = finnotech.GetToken(code)
			if err != nil {
				ct.JSON(http.StatusBadRequest, gin.H{
					"message": "Failed to verify token2",
					"status": err.Error(),
				})
				return
			}

			val = PhoneVerifications[v.Phone]
			val.FinnotechToken = finnotech.Token
			PhoneVerifications[v.Phone] = val
	
		} else {
			finnotech.Token = PhoneVerifications[v.Phone].FinnotechToken
		}

		err := finnotech.VerifyNid(v.Ssn, v.Birthdate, v.FName,v.LName, v.Gender)
		if err != nil {
			ct.JSON(http.StatusBadRequest, gin.H{
				"message": "failed",
				"status": err.Error(),
			})
			return
		}
	
		fmt.Println("store user db")
		//TODO password
		password := strconv.Itoa(util.RandInt(1000000, 9999999))
		userInfo := &model.UserInfo{
			User: model.User{
				Username: v.Username,
				Password: password,
			},
			Phone: v.Phone,
			FaName: v.FName,
			FaFamily: v.LName,
			NationalID: v.Ssn,
			Gender: v.Gender,
		}
		db := database.GetExec()
	
		if db.Create(userInfo).Error == nil {
			token, _ :=jwt.CreateToken(userInfo.UserID);


			jsonBody,_ := json.Marshal(map[string]string{
				"uid": userInfo.User.UID.String(),
			})
			client := &http.Client{
				Timeout: time.Second * 10,
			}
			req, _ := http.NewRequest("POST", CoreNewUser, bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type","application/json")
			res, err := client.Do(req)
			if err != nil  {
				fmt.Println("core new user failed")
				return
			}
			if res.StatusCode != 200 {
				fmt.Println("uid exists new user failed")
				return
			}

			ct.JSON(http.StatusOK, gin.H{
				"token": token,
			})
			
			go func ()  {
				message := password
				notification.SendSMS(message,[]string{v.Phone}, "password")
				delete(PhoneVerifications, v.Phone)
			}()
			return
		}
	}
	
	ct.JSON(http.StatusBadRequest, gin.H{
		"message": "Failed",
		"status": "duplicateUsername",
	})


}


func ValidateUser(ct *gin.Context) {
	userID := ct.MustGet("userId").(int)	
	ct.JSON(http.StatusOK, gin.H{
		"user": userID,
	})


}

func GetUserWithToken(ct *gin.Context) {

	userID := ct.MustGet("userId").(int)
	userInfo := model.UserInfo{}
	db := database.GetExec()
	result := db.Preload("User").First(&userInfo, "user_id = ?", userID)
	fmt.Println(userInfo)
	if result.Error == nil && result.RowsAffected == 1 {
		//TODO real name 

		ct.JSON(http.StatusOK, gin.H{
			"firstName": userInfo.FaName, 
			// userInfo.FaName,
			"lastName": userInfo.FaFamily,
			"verified": userInfo.User.IsVerified(),
			// userInfo.FaFamily,
			"username": userInfo.User.Username,
			"userLevel": 1,
			"gender": userInfo.Gender,
			"phone": userInfo.Phone,
			"uid": userInfo.User.UID,
		})
		return
	}

	ct.JSON(http.StatusInternalServerError, "Failed")

}


func GetUserId(ct *gin.Context) (int, error) {
	if tokenString := ExtractToken(ct); tokenString != "" {
		claims, err := jwt.VerifyToken(tokenString)
		if err != nil {
			return 0, err
		}
		return int(claims["user_id"].(float64)), nil
	}

	
	return 0, fmt.Errorf("token not found")
}

func LogOut(ct *gin.Context) {
	if tokenString := ExtractToken(ct); tokenString != ""{
		jwt.InvalidateToken(tokenString)
		ct.JSON(http.StatusOK, "Successful")
		return

	}
	ct.JSON(http.StatusBadRequest,"")
}

func checkCode(v interface{}, st SMSType) bool {
	switch v := v.(type) {
	case SignUpInputs:

		if val, ok := PhoneVerifications[v.Phone]; ok && val.TempCode == v.TempCode  && val.Type == st {
			return true
		}
	case LoginVerifyInputs:
		if val, ok := PhoneVerifications[v.Username]; ok && val.TempCode == v.TempCode && val.VerifyCode == v.VerifyCode && val.Type == st && time.Now().Before(val.Expire){
			return true
		}
	case ForgotPasswordInput:
		if val, ok := PhoneVerifications[v.Phone]; ok && val.TempCode == v.TempCode && val.VerifyCode == v.VerifyCode && val.Type == st && time.Now().Before(val.Expire){
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
			VerifyCode: util.RandInt(10000, 99999),
			TempCode: util.RandInt(10000, 99999),
			Type: Login_sms,
			UserID: int(user.ID),
			Expire: time.Now().Add(time.Minute * 2),

	
		}
		PhoneVerifications[credentials.Username] = v

		var userInfo model.UserInfo

		result = db.Where("user_id = ? ", user.ID).First(&userInfo)
		if result.Error == nil && result.RowsAffected == 1   {
			ct.JSON(http.StatusOK, gin.H{
				"tempCode": v.TempCode,
			})
			message := strconv.Itoa(v.VerifyCode)
			go notification.SendSMS(message,[]string{userInfo.Phone}, "code")
	
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

func SendForgotPasswordSMS(ct *gin.Context)  {
	var in PhoneInput
	if err := ct.ShouldBind(&in) ; err != nil {
		ct.JSON(http.StatusBadRequest, gin.H{
			"message": "Wrong Input",
		})
		return
	}
	v := VerificationCodes{
		VerifyCode: util.RandInt(10000, 99999),
		TempCode: 12345,
		// util.RandInt(10000, 99999),
		Type: Forgot_Password_sms,
		Expire: time.Now().Add(2 * time.Minute),

	}
	PhoneVerifications[in.Phone] = v
	message := strconv.Itoa(v.VerifyCode)
	go notification.SendSMS(message, []string{in.Phone}, "forgotPassword")

	ct.JSON(http.StatusOK, gin.H{
			"message": "successful",
	})

}

func VerifyForgotPassword(ct *gin.Context) {
	var in ForgotPasswordInput

	if err := ct.ShouldBind(&in) ; err != nil {
		fmt.Println(in)
		ct.JSON(http.StatusBadRequest, gin.H{
			"message": "Wrong Input",
		})
		return
	}

	if checkCode(in, Forgot_Password_sms) {
		password := strconv.Itoa(util.RandInt(100000, 999999))
		db := database.GetExec()
		var userInfo model.UserInfo

		if db.Preload("User").Last(&userInfo, "phone = ?", in.Phone).Error != nil {
			ct.JSON(http.StatusBadRequest, gin.H{
				"message": "phone not found",
			})
			return
		}
		u := model.User{
			Password: password,
		}
		if db.Model(&u).Where("id = ?", userInfo.UserID).Updates(&u).Error != nil {
			ct.JSON(http.StatusBadRequest, gin.H{
				"message": "server error",
			})
			return
		}
		message := password
		go notification.SendSMS(message, []string{in.Phone}, "password")
		token, _ := jwt.CreateToken(userInfo.UserID);
		delete(PhoneVerifications, in.Phone)

		ct.JSON(http.StatusOK, gin.H{
			"token": token,
		})
		return

	}
	ct.JSON(http.StatusBadRequest, gin.H{
		"message": "wrong code",
	})

}


func ChangePassword(ct *gin.Context) {
	var in ChangePasswordInput
	userId := ct.MustGet("userId").(int)
	if err := ct.ShouldBind(&in) ; err != nil {
		ct.JSON(http.StatusBadRequest, gin.H{
			"message": "Wrong Input",
		})
		return
	}
	db := database.GetExec()
	u := model.User{}
	if err := db.First(&u, userId).Error; err != nil {
		ct.JSON(http.StatusBadRequest, gin.H{
			"message": "user not found",
		})
		return
	}
	if err := u.CheckPassword(in.OldPassword); err != nil {
		ct.JSON(http.StatusBadRequest, gin.H{
			"message": "wrong password",
		})
		return
	}

	user := model.User{}
	user.Password = in.NewPassword
	user.ID = u.ID 
	if err := db.Debug().Model(&user).Updates(&user).Error; err != nil {
		ct.JSON(http.StatusBadRequest, gin.H{
			"message": "server error",
		})
		return
	}

	ct.JSON(http.StatusOK, gin.H{
			"message": "successful",
		})



}

func ExtractToken (ct *gin.Context) string {
	bearToken := ct.GetHeader("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
	   return strArr[1]
	}
	return ""
}


