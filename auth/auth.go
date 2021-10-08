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
	"example.com/models/jwt"
	"example.com/models/model"
	"example.com/models/notification"
	"example.com/models/util"
	"github.com/gin-gonic/gin"
)


type SignUpPhone struct {
	Phone string `json:"phone" form:"phone" binding:"required"`
}

type VerificationCodes struct {
	VerifyCode	int		`json:"verifyCode"`
	TempCode	int		`json:"tempCode"`
	Type		SMSType
	UserID		int
	TrackID		string `json:"trackID"`
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

const (
	FinnotechAuthorize 			string = "https://apibeta.finnotech.ir/dev/v2/oauth2/authorize"
	FinnotechVerifySMS 			string = "https://apibeta.finnotech.ir/dev/v2/oauth2/verify/sms"
	FinnotechGetToken 			string = "https://apibeta.finnotech.ir/dev/v2/oauth2/token"
	FinnotechNidVerification 	string = "https://apibeta.finnotech.ir/facility/v2/clients/REDACTED/users/"
)

const (
	CoreNewUser string = "http://192.168.160.207:22216/newuser"
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
	var phoneInput SignUpPhone
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

		jsonBody, _ := json.Marshal(map[string]string{
			"mobile": v.Phone,
			"otp":	strconv.Itoa(v.VerifyCode),
			"nid":	v.Ssn,
			"trackId":	PhoneVerifications[v.Phone].TrackID,	

		})
		client := &http.Client{
			Timeout: time.Second * 10,
		}
		req, _ := http.NewRequest("POST", FinnotechVerifySMS, bytes.NewBuffer(jsonBody))
		req.Header.Add("Authorization","bearer " + os.Getenv("FINNOTECH_AUTHORIZE_STRING"))
		req.Header.Set("Content-Type","application/json")
		res, err := client.Do(req)

		if err != nil {
			ct.JSON(http.StatusBadRequest, gin.H{
				"message": "Failed",
			})
			return
		}
		r := struct{
			Result struct {
				Code string `json:"code"`
			} `json:"result"`
			Status string `json:"status"`
		}{} 
		err = json.NewDecoder(res.Body).Decode(&r)
		if err != nil || r.Status != "DONE" {
			ct.JSON(http.StatusBadRequest, gin.H{
				"message": "Failed",
			})
			return
		}

		jsonBody, _ = json.Marshal(map[string]string{
			"grant_type": "authorization_code",
			"code": r.Result.Code,
			"auth_type": "SMS",
			"redirect_uri": "REDACTED",

		})
		req, _ = http.NewRequest("POST", FinnotechGetToken, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type","application/json")
		req.Header.Add("Authorization","Basic " + os.Getenv("FINNOTECH_AUTHORIZE_STRING"))
		res, err = client.Do(req)

		if err != nil {
			fmt.Println(err.Error())
			ct.JSON(http.StatusBadRequest, gin.H{
				"message": "Failed",
			})
			return
		}
		r1 := struct {
			Result struct {
				ClientID     string   `json:"clientId"`
				Scopes       []string `json:"scopes"`
				LifeTime     string   `json:"lifeTime"`
				Type         string   `json:"type"`
				CreationDate string   `json:"creationDate"`
				UserID       string   `json:"userId"`
				AuthType     string   `json:"auth_type"`
				RefreshToken string   `json:"refreshToken"`
				Bank         string   `json:"bank"`
				Value        string   `json:"value"`
			} `json:"result"`
			Status string `json:"status"`
		}{}
		fmt.Println(r1)
		err = json.NewDecoder(res.Body).Decode(&r1)
		if err != nil || r1.Status != "DONE" {
			ct.JSON(http.StatusBadRequest, gin.H{
				"message": "Failed",
			})
			return
		}


		req, _ = http.NewRequest("GET", FinnotechNidVerification + v.Ssn + "/sms/nidVerification", nil)
		q := req.URL.Query()
		//TODO read env
		q.Add("birthDate", v.Birthdate)
		q.Add("gender", v.Gender)
		q.Add("firstName", v.FName)
		q.Add("lastName", v.LName)


		req.URL.RawQuery = q.Encode()
		req.Header.Add("Authorization","bearer " + r1.Result.Value)
		res, err = client.Do(req)
		if err != nil || res.StatusCode != 200 {
			ct.JSON(http.StatusBadRequest, gin.H{
				"message": "Failed",
			})
			return
		}

		r2 := struct {
			TrackID string `json:"trackId"`
			Result  struct {
				InvalidField        int     `json:"invalidField"`
				NationalCode        string  `json:"nationalCode"`
				BirthDate           string  `json:"birthDate"`
				FirstName           string  `json:"firstName"`
				FirstNameSimilarity float64 `json:"firstNameSimilarity"`
				LastName            string  `json:"lastName"`
				LastNameSimilarity  float64 `json:"lastNameSimilarity"`
				Gender              string  `json:"gender"`
				GenderSimilarity    int     `json:"genderSimilarity"`
				DeathStatus         string  `json:"deathStatus"`
			} `json:"result"`
			Status string `json:"status"`
		}{}
		err = json.NewDecoder(res.Body).Decode(&r2)

		if err != nil || r2.Status != "DONE" || r2.Result.GenderSimilarity < 100 || r2.Result.FirstNameSimilarity < 100 || r2.Result.LastNameSimilarity < 100  {
			ct.JSON(http.StatusBadRequest, gin.H{
				"message": "information not match",
			})
			return 
		}

		fmt.Println("dbbbbbbbbbbbbbb")
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
		}
		db := database.GetExec()

		if db.Create(userInfo).Error == nil {
			token, _ :=jwt.CreateToken(userInfo.UserID);
			ct.JSON(http.StatusOK, gin.H{
				"token": token,
			})
			go func ()  {
				message := "password: " + password
				notification.SendSMS(message,[]string{v.Phone})
				delete(PhoneVerifications, v.Phone)
				jsonBody,_ := json.Marshal(map[string]string{
					"uid": userInfo.User.UID.String(),
				})
				req, _ = http.NewRequest("POST", CoreNewUser, bytes.NewBuffer(jsonBody))
				req.Header.Set("Content-Type","application/json")
				res, err = client.Do(req)
				if err != nil || res.StatusCode != 200 {
					fmt.Println("core failed", res.StatusCode)
				}
			}()
			return
		}
	}
	
	ct.JSON(http.StatusBadRequest, gin.H{
		"message": "Failed",
	})


}


func GetUser(ct *gin.Context) {
	if tokenString := ExtractToken(ct); tokenString != ""{
		if claims, err := jwt.VerifyToken(tokenString); err == nil {
			ct.JSON(http.StatusOK, gin.H{
				"user": claims["user_id"],
			})
			return
		} 

	}

	ct.JSON(http.StatusUnauthorized,"Unauthorized.")
}

func GetUserWithToken(ct *gin.Context) {
	if tokenString := ExtractToken(ct); tokenString != ""{
		if claims, err := jwt.VerifyToken(tokenString); err == nil {
			userID := claims["user_id"]
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
					"uid": userInfo.User.UID,
				})
				return
			}
			ct.JSON(http.StatusInternalServerError, "Failed")


		} 

	}

	ct.JSON(http.StatusUnauthorized,"Unauthorized.")
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
			VerifyCode: util.RandInt(10000, 99999),
			TempCode: util.RandInt(10000, 99999),
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




func ExtractToken (ct *gin.Context) string {
	bearToken := ct.GetHeader("Authorization")
	strArr := strings.Split(bearToken, " ")
	if len(strArr) == 2 {
	   return strArr[1]
	}
	return ""
}