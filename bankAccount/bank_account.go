package bankaccount

import (
	"fmt"
	"strings"

	"example.com/models/database"
	"example.com/models/finnotech"
	"example.com/models/model"
)


const (
	ShebaFromCardAddress 				= "https://apibeta.finnotech.ir/facility/v2/clients/REDACTED/cardToIban"
	BankAccountOwnerVerificationAddress =  ""
)

type BankAccount struct {
	Card		string
	Iban		string
	BankCode 	string
	BankName	string
}


type CardToShebaResponse struct {
	TrackID string `json:"trackId"`
	Result  struct {
		Iban               string `json:"IBAN"`
		BankName           string `json:"bankName"`
		Deposit            string `json:"deposit"`
		Card               string `json:"card"`
		DepositStatus      string `json:"depositStatus"`
		DepositDescription string `json:"depositDescription"`
		DepositComment     string `json:"depositComment"`
		DepositOwners      []struct {
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		} `json:"depositOwners"`
	} `json:"result"`
	Status string `json:"status"`
}




func (b *BankAccount)CheckBankAccountInput() bool {
	if b.Card == "" || b.Iban == "" || b.BankName == "" || b.BankCode == "" || len(b.Card) != 16 || len(b.Iban) != 26 || b.Iban[0:2] != "IR" {
		return false
	}
	return true
}

func (b *BankAccount)VerifyBankAccount(userId int) error {
	f := finnotech.NewFinnotech()
	return f.VerifyBankInfoMatch(b.Card, b.Iban, userId)

}

func (b *BankAccount)CheckIsBankAccountOwner(userId int) error {
	userInfo := model.UserInfo{}
	db := database.GetExec()
	if res := db.First(&userInfo, "user_id", userId); res.Error!= nil{
		return fmt.Errorf("UserNotFound")
	}
	finnotech.NewFinnotech()
	return nil
}

func (b *BankAccount)ToString() string {
	return b.BankName + "-" + b.BankCode  + "-" + b.Card + "-" + b.Iban
}


func BuildFromString(s string) *BankAccount {
	arr := strings.Split(s,"-")
	return &BankAccount{
		BankName: arr[0],
		BankCode: arr[1],
		Card: arr[2],
		Iban: arr[3],
	}
}