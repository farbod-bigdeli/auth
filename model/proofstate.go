package model

import (
	// "time"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type ProofState struct {
	gorm.Model
	UID				uuid.UUID	`gorm:"type:uuid;default:gen_random_uuid();"`	
	State			string		`gorm:"not null"`

}

const(
	NotVerifiedState 	int = 1
	VerifiedState		int = 2
	FullVerifiedState 	int = 3
)