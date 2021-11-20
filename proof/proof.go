package proof

import (
	"fmt"
	"net/http"
	"strconv"

	"example.com/models/bankAccount"
	"example.com/models/database"
	"example.com/models/model"
	"example.com/models/util"


	"github.com/gin-gonic/gin"
)








func Store(c *gin.Context)  {
	db := database.GetExec();
	userID := c.MustGet("userId").(int)
	proofTypeString := c.Request.FormValue("type")
	proofType, err := strconv.Atoi(proofTypeString)
	fmt.Printf("%T   %v", proofType, proofType )

	if err != nil {
		c.JSON(http.StatusBadRequest, "Invalid type")
		return
	}
	proof := model.UserProof{
		UserID: userID,
	}
	switch proofType {

	case model.CommitmentProofType:
		proof.ProofTypeID = model.CommitmentProofType
		proof.ProofStateId = model.FullVerifiedState
		proof.Path, err = util.UploadPhoto(*c, "commitment/")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message":"Invalid photo",
				"status": err.Error(),
			})
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
		err := bankAccount.VerifyBankAccount(userID)
		//TODO check owner
		// err1 := bankAccount.CheckIsBankAccountOwner(userID)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "verification failed",
				"status": err.Error(),
			})
			return
		}
		proof.ProofStateId = model.FullVerifiedState
		proof.Path = bankAccount.ToString()
		p := model.UserProof{}
		if db.Last(&p, "path = ? AND user_id = ?", proof.Path, userID).Error == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": "card exists",
				"status": "duplicateCard",
			})
			return
		}
	default :
		c.JSON(http.StatusBadRequest, "Invalid type")
		return
	
	}

	if db.Create(&proof).Error == nil {
		c.JSON(http.StatusOK, "Successful")
		return

	}
	c.JSON(http.StatusForbidden, "Failed")
}


func GetBankAccount(ct *gin.Context) {
	userID := ct.MustGet("userId")
	db := database.GetExec()
	var p []model.UserProof
	db.Find(&p,"user_id = ? AND proof_type_id = ?", userID, model.BankAccountProofType)

	var res []*bankaccount.BankAccount = []*bankaccount.BankAccount{}
	for _, v := range p {
		res = append(res, bankaccount.BuildFromString(v.Path))
	}
	ct.JSON(http.StatusOK, res)
}





