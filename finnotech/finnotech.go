package finnotech

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"example.com/models/database"
	"example.com/models/model"
	"github.com/mavihq/persian"
)

type Finnotech struct {
	Client *http.Client
	Token string
}
var ClientCredentialToken string = ""
var ClientCredentialRefreshToken string = ""
var ClientCredentialTokenRefreshTime int = 0

const (
	FinnotechAuthorize 			string = "https://apibeta.finnotech.ir/dev/v2/oauth2/authorize"
	FinnotechVerifySMS 			string = "https://apibeta.finnotech.ir/dev/v2/oauth2/verify/sms"
	FinnotechGetToken 			string = "https://apibeta.finnotech.ir/dev/v2/oauth2/token"
	FinnotechNidVerification 	string = "https://apibeta.finnotech.ir/facility/v2/clients/REDACTED/users/"
	FinnotechShebaFromCard		string = "https://apibeta.finnotech.ir/facility/v2/clients/REDACTED/cardToIban"
)

type FinnotechResponse struct {
	Result struct {
		Value        string   `json:"value"`
		Scopes       []string `json:"scopes"`
		LifeTime     int      `json:"lifeTime"`
		CreationDate string   `json:"creationDate"`
		RefreshToken string   `json:"refreshToken"`
	} `json:"result"`
	Status string `json:"status"`
}

var do sync.Once

func getToken() {
	jsonBody, _ := json.Marshal(map[string]string{
		"grant_type": "client_credentials",
		"nid": "REDACTED",
		"scopes": "facility:card-to-iban:get",

	})
	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	req, _ := http.NewRequest("POST", "https://apibeta.finnotech.ir/dev/v2/oauth2/token", bytes.NewBuffer(jsonBody))
	req.Header.Add("Authorization","Basic " + os.Getenv("FINNOTECH_AUTHORIZE_STRING"))
	req.Header.Set("Content-Type","application/json")
	res, err := client.Do(req)
	if err != nil {
		panic("failed to get finnotech token")
	}
	r := FinnotechResponse{}
	err = json.NewDecoder(res.Body).Decode(&r)
	if err != nil {
		panic("failed to get finnotech token")

	}
	ClientCredentialToken = r.Result.Value
	ClientCredentialRefreshToken = r.Result.RefreshToken
	ClientCredentialTokenRefreshTime = r.Result.LifeTime
	go func(){
		ticker := time.NewTicker(time.Millisecond * time.Duration(ClientCredentialTokenRefreshTime - 10000))
		done := make(chan bool)
		for {
			select{
			case <- ticker.C:
			jsonBody, _ := json.Marshal(map[string]string{
				"grant_type": "refresh_token",
				"token_type":"CLIENT-CREDENTIAL",
				"refresh_token": ClientCredentialRefreshToken,
		
			})
			client := &http.Client{
				Timeout: 10 * time.Second,
			}
			req, _ := http.NewRequest("POST", "https://apibeta.finnotech.ir/dev/v2/oauth2/token", bytes.NewBuffer(jsonBody))
			req.Header.Add("Authorization","Basic " + os.Getenv("FINNOTECH_AUTHORIZE_STRING"))
			req.Header.Set("Content-Type","application/json")
			res, err := client.Do(req)
			if err != nil {
				panic("failed to get finnotech token")
			}
			r := FinnotechResponse{}
			err = json.NewDecoder(res.Body).Decode(&r)
			if err != nil {
				panic("failed to get finnotech token")
		
			}
			ClientCredentialToken = r.Result.Value
			ClientCredentialRefreshToken = r.Result.RefreshToken
			fmt.Println("new token:  ",ClientCredentialToken)

			case <-done:
				return
		}
		}
	}()
}

func NewFinnotech() Finnotech {
	do.Do(getToken)
	return Finnotech{
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (f Finnotech) VerifySMS(phone, otp, nid, trackId string) (string, error){
	jsonBody, _ := json.Marshal(map[string]string{
		"mobile": phone,
		"otp":	otp,
		"nid":	nid,
		"trackId":	trackId,	

	})

	req, _ := http.NewRequest("POST", FinnotechVerifySMS, bytes.NewBuffer(jsonBody))
	req.Header.Add("Authorization","Basic " + os.Getenv("FINNOTECH_AUTHORIZE_STRING"))
	req.Header.Set("Content-Type","application/json")
	start := time.Now()
	res, err := f.Client.Do(req)
	fmt.Println(time.Since(start))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	r := struct{
		Result struct {
			Code string `json:"code"`
		} `json:"result"`
		Status string `json:"status"`
	}{} 
	err = json.NewDecoder(res.Body).Decode(&r)
	if err != nil || r.Status != "DONE" {
		return "", fmt.Errorf("wrongCode")
	}
	if r.Status == "VALIDATION_ERROR" {
		return "", fmt.Errorf("wrongNid")

	}

	fmt.Println("first done")
	return r.Result.Code, nil
}

func (f *Finnotech)GetToken(code string) (error) {
	jsonBody, _ := json.Marshal(map[string]string{
		"grant_type": "authorization_code",
		"code":code,
		"auth_type": "SMS",
		"redirect_uri": "REDACTED",
	
	})
	req, _ := http.NewRequest("POST", FinnotechGetToken, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type","application/json")
	req.Header.Add("Authorization","Basic " + os.Getenv("FINNOTECH_AUTHORIZE_STRING"))
	start := time.Now()
	res, err := f.Client.Do(req)
	fmt.Println(time.Since(start))
	
	if err != nil {
		return fmt.Errorf("server error")
	}
	defer res.Body.Close()

	r := struct {
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
	err = json.NewDecoder(res.Body).Decode(&r)
	if err != nil || r.Status != "DONE" {
		return fmt.Errorf("failed to get token")
	}
	fmt.Println("second done")
	f.Token = r.Result.Value
	return nil
}


func (f Finnotech)VerifyNid(ssn, birthdate,fName,lName, gender string) (error) {
	req, _ := http.NewRequest("GET", FinnotechNidVerification + ssn + "/sms/nidVerification", nil)
	q := req.URL.Query()
	//TODO read env
	q.Add("birthDate", birthdate)
	q.Add("gender", gender)
	q.Add("firstName", fName)
	q.Add("lastName", lName)
	
	
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Authorization","bearer " + f.Token)
	start := time.Now()
	res, err := f.Client.Do(req)
	fmt.Println(time.Since(start))
	if err != nil || res.StatusCode != 200 {
		return fmt.Errorf("server error")
	}
	defer res.Body.Close()

	r := struct {
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
	err = json.NewDecoder(res.Body).Decode(&r)
	
	if err != nil || r.Status != "DONE" || r.Result.GenderSimilarity < 100 || r.Result.FirstNameSimilarity < 100 || r.Result.LastNameSimilarity < 100  {
		return fmt.Errorf("wrongIdentity")
	}
	return nil
}

func(f Finnotech)VerifyBankInfoMatch(cardNumber, iban string, userId int) error{
	req, _ := http.NewRequest("GET", FinnotechShebaFromCard, nil)
	q := req.URL.Query()
	q.Add("card", cardNumber)
	q.Add("version", "2")
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Authorization","Bearer " + ClientCredentialToken)
	res, err := f.Client.Do(req)
	if err != nil {
		return fmt.Errorf("Failed")
	}
	r := struct {
		TrackID string `json:"trackId"`
		Result  struct {
			Iban               string `json:"IBAN"`
			BankName           string `json:"bankName"`
			Deposit            string `json:"deposit"`
			Card               string `json:"card"`
			DepositStatus      string `json:"depositStatus"`
			DepositDescription string `json:"depositDescription"`
			DepositComment     string `json:"depositComment"`
			DepositOwners      string `json:"depositOwners"`
		} `json:"result"`
		Status string `json:"status"`
	}{}
	// b,_ := io.ReadAll(res.Body)
	// fmt.Println(string(b))
	json.NewDecoder(res.Body).Decode(&r)

	if res.StatusCode != 200 || r.Status != "DONE" {
		return fmt.Errorf("CardNOTValid")
	}
	if r.Result.DepositStatus != "02" {
		return fmt.Errorf("InvalidDepsit")
	}
	if r.Result.Iban != iban {
		return fmt.Errorf("NotMatch")

	}
	str := strings.Split(r.Result.DepositOwners, "/")[0]
	depositeOwner := persian.FixArabic(str[:len(str)-1])
	userInfo := model.UserInfo{}
	db := database.GetExec()
	result := db.Where("user_id = ? ", userId).First(&userInfo)
	if result.Error != nil {
		return fmt.Errorf("UserNotFound")
	}
	fullname := persian.FixArabic(fmt.Sprintf("%s %s", userInfo.FaName ,userInfo.FaFamily))

	if fullname != depositeOwner {
		return fmt.Errorf("CardNotForOwner")
	}

	return nil
}