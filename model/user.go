package model

import (
	"example.com/models/database"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UID        	uuid.UUID 	`gorm:"type:uuid;default:gen_random_uuid();"`
	Username	string		`gorm:"not null;unique"`
	Password 	string		`gorm:"not null;"`

}


func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	hash, err :=  bcrypt.GenerateFromPassword([]byte(u.Password), 10)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return


}

func (u *User) BeforeUpdate(tx *gorm.DB) (err error) {
	hash, err :=  bcrypt.GenerateFromPassword([]byte(u.Password), 10)
	if err != nil {
		return err
	}
	u.Password = string(hash)
	return


}

func (u *User) CheckPassword(password string) error {
	bytePassword := []byte(password)
	byteHashedPassword := []byte(u.Password)
	return bcrypt.CompareHashAndPassword(byteHashedPassword, bytePassword)
}

func (u User) IsVerified() bool {
	var commitment, bankAccount UserProof
	db := database.GetExec()
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := db.First(&commitment, "user_id = ? AND proof_type_id = ? AND proof_state_id = ?", u.ID, CommitmentProofType, FullVerifiedState ).Error; err != nil {
			return err
		}
		if err := db.First(&bankAccount, "user_id = ? AND proof_type_id = ? AND proof_state_id = ?", u.ID, BankAccountProofType, FullVerifiedState).Error; err != nil {
			return err
		}
		return nil

	})
	if err != nil {
		return false
	}
	return err == nil
}
