package model

import (
	// "time"

	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type UserInfo struct {
	gorm.Model
	ID        				int 		`gorm:"primary_key;"`
	UID						uuid.UUID	`gorm:"type:uuid;default:gen_random_uuid();"`	
	UserID					int			
	User					User		
	Name 					string				
	Family 					string			
	FaName 					string			
	FaFamily 				string		
	Email 					string		
	FaState					string		
	Phone					string		
	Tel						string			
	NationalID				string		
	BirthCertificateID		string		
	Nationality	 			string		
	

}