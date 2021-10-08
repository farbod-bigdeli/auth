package proof

import (
	"fmt"
	"net/http"
	"strconv"

	"example.com/models/auth"
	"example.com/models/bankAccount"
	"example.com/models/database"
	"example.com/models/jwt"
	"example.com/models/model"
	"example.com/models/util"


	"github.com/gin-gonic/gin"
)



type BankAccount struct {
	Card		string
	Iban		string
	BankCode 	string
	BankName	string
}




func Store(c *gin.Context)  {

	tokenString := auth.ExtractToken(c);
	if  tokenString == "" {
		c.JSON(http.StatusBadRequest, "Token not found")
		return
	}
	 claims, err := jwt.VerifyToken(tokenString)
	 if  err != nil {
		c.JSON(http.StatusBadRequest, "Invalid token")
		return
	}
	proofTypeString := c.Request.FormValue("type")
	proofType, err := strconv.Atoi(proofTypeString)
	fmt.Printf("%T   %v", proofType, proofType )

	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid type")
		return
	}
	userID := int(claims["user_id"].(float64))
	proof := model.UserProof{
		UserID: userID,
	}
	switch proofType {

	case model.CommitmentProofType:
		proof.ProofTypeID = model.CommitmentProofType
		proof.ProofStateId = model.FullVerifiedState
		proof.Path, err = util.UploadPhoto(*c, "commitment/")
		if err != nil {
			c.JSON(http.StatusBadRequest, "Invalid photo")
			return
		}
	case model.BankAccountProofType:
		proof.ProofTypeID = model.BankAccountProofType
		bankAccount := bankaccount.BankAccount{
			Card: c.Request.FormValue("cardNumber"),
			Iban: c.Request.FormValue("shebaNumber"),
			BankCode: c.Request.FormValue("bankCode"),
			BankName: c.Request.FormValue("bankName"),
		}
		
		if !bankAccount.CheckBankAccountInput() {
			c.JSON(http.StatusBadRequest, "Invalid input")
			return
		}
			//TODO verify bank account
			// if !bankAccount.VerifyBankAccount() {
			// 	c.JSON(http.StatusBadRequest, "Invalid card or sheba")
			// 	return
			// }
		proof.ProofStateId = model.FullVerifiedState
		proof.Path = bankAccount.ToString()

	default :
		c.JSON(http.StatusBadRequest, "Invalid type")
		return
	
	}

	db := database.GetExec();
	if db.Create(&proof).Error == nil {
		c.JSON(http.StatusOK, "Successful")
		return

	}
	c.JSON(http.StatusForbidden, "Failed")
}






