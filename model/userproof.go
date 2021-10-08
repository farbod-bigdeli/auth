package model

import (
	// "time"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type UserProof struct {
	gorm.Model
	UID					uuid.UUID		`gorm:"type:uuid;default:gen_random_uuid();"`	
	UserID				int				`gorm:"not null;"`
	User				User			`gorm:"not null"`
	ProofStateId		int				`gorm:"not null;default:1;"`
	ProofState			ProofState		`gorm:"not null"`
	ProofTypeID			int				`gorm:"not null"`
	ProofType			ProofType		`gorm:"not null"`
	Path 				string			`gorm:"not null"`

}

