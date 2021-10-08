package bankaccount

import (
	"net/http"
	"encoding/json"
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

func (b *BankAccount)VerifyBankAccount() bool {

	client := &http.Client{}

	req, _ := http.NewRequest("GET", ShebaFromCardAddress, nil)
	q := req.URL.Query()
	q.Add("card", b.Card)
	req.URL.RawQuery = q.Encode()
	req.Header.Add("Authorization","bearer " + "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0b2tlbklkIjoiMGM4YjBhYzYtODAwNy00ZjVjLTk1MWYtNTk3NDFiZTM5ZTMyIiwicmVmcmVzaFRva2VuIjoiVFg5MzVaUTN6c0VoTDVqQ0FHRnpXTDNhaktlTDM4TklnQU0xZFFpenVXMGVsNkZnNUFGUHVOSlF3S01YV2toYXRlSGx1cWR4MTBRSVd1cFE2VFR2UlV0V2VTVG9JNmFTOGVXSnBIMTBiWlozWkxoelJ2bUZIRU1sbzdnM3I2NVFBSmx5Z0ZEQ3gwRmxKdlVMRUJZV1U2VFAzaTUybzF0RnJ5NlVwampwUk1KZlVkOXBUekZ5bDZKd0R2bkhkakxlTm1mRDNTVW10bzI5TDh1THZLaW10dVY2NzhnYnM4eTJyZmR6eHVwQkM2QTAydnlWRm9Ecnl5UlFSRGFsaEVhdiIsImNyZWF0aW9uRGF0ZSI6IjE0MDAwNzExMDg1NTI5IiwibGlmZVRpbWUiOjg2NDAwMDAwMCwiY2xpZW50SWQiOiJUYWJhQ2hhaW4xIiwidXNlcklkIjoiMDAyMjIxNjI1MSIsImFjdGl2ZSI6dHJ1ZSwic2NvcGVzIjpbImZhY2lsaXR5OmNhcmQtdG8taWJhbjpnZXQiXSwidHlwZSI6IkNMSUVOVC1DUkVERU5USUFMIiwiY2xpZW50SW5mbyI6Ilg3VzZoUUF5TnR0TXR4YUNqVFJIMHV3ei9UZmFTY0JUamNPK0JRK3c0aEc0ZEJ1OFdjMWVIbmdBVHdYdGdRZlEiLCJiYW5rIjoiMDYyIiwiaWF0IjoxNjMzMjM4NzI5LCJleHAiOjE2MzQxMDI3Mjl9.E4sYWAgGp-pFYll4wAKcV_4IjYd03o0dLEXMK1cpvjODpEnKRSYlswk9SRnTKhAnVFWyCHEw5ZEFGOcENvGQAZpCQZ4_0zdL-qfhM3sKZcH-hG4QU8uyvf1jRBdwwmvZ2gM6UckrP4Aga_2zJ2y0VoVsfRjHir8NV1TG1LPuAOrL-TosHU7-AN2Dz37BI2w7EMTEF1o0JYDhfyg1SBhooEZR1E7J2-31rTlzVFU6NmDTyGUeL4j53JNSEiLykMPO0GQGIsE5FQpQklu6ezdmTBsxEzFdtDZUiel46UnohK_xZKbC-Tas5O2syiwSzKJ1NOcsIA11ycnnXkasSnYqyQ")
	res, err := client.Do(req)
	if err != nil {
		return false
	}
	var r CardToShebaResponse
	json.NewDecoder(res.Body).Decode(&r)

	if res.StatusCode == 200 && r.Status == "DONE" && r.Result.DepositStatus == "02" && r.Result.Iban == b.Iban {
		return true
	}

		
	return false

}

func (b *BankAccount)checkIsBankAccountOwner() {

}

func (b *BankAccount)ToString() string {
	return b.BankName + "-" + b.BankCode  + "-" + b.Card + "-" + b.Iban
}