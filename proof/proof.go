package proof

import (
	"fmt"
	"net/http"
	"strconv"

	"example.com/models/auth"
	"example.com/models/database"
	"example.com/models/jwt"
	model "example.com/models/models"
	"example.com/models/util"
	"github.com/gin-gonic/gin"
)


type ProofType int

type ProofState int


const(
	Commitment 		ProofType = 1
	BankAccount 	ProofType = 2

)

const(
	NotVerified 	ProofState = 1
	Verified 		ProofState = 2

)


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
	proof := model.UserProof{
		UserID: int(claims["user_id"].(float64)),
	}
	pt := ProofType(proofType)
	switch pt {
	case Commitment:
		proof.ProofTypeID = int(Commitment)
		proof.Path, err = util.UploadPhoto(*c, "commitment/")
		if err != nil {
			c.JSON(http.StatusBadRequest, "Invalid photo")
			return
		}
	case BankAccount:
		proof.ProofTypeID = int(BankAccount)
		cardNumber := c.Request.FormValue("cardNumber")
		bankName := c.Request.FormValue("bankName")
		shebaNumber := c.Request.FormValue("shebaNumber")

		if cardNumber == "" || bankName == "" || shebaNumber == "" {
			c.JSON(http.StatusBadRequest, "Invalid input")
			return
		}
		proof.Path = bankName + "-" + cardNumber + "-" + shebaNumber

	default :
		c.JSON(http.StatusBadRequest, "Invalid type")
		return
	
	}

	db := database.GetExec();
	if db.Create(&proof).Error == nil {
		c.JSON(http.StatusOK, "Successful")

	}
	c.JSON(http.StatusForbidden, "Failed")
}