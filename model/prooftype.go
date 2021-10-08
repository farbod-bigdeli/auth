package model

import (
	// "time"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type ProofType struct {
	gorm.Model
	UID					uuid.UUID		`gorm:"type:uuid;default:gen_random_uuid();"`	
	Description			string			`gorm:"not null"`
}



const(
	CommitmentProofType 	int = 1
	BankAccountProofType 	int = 2

)