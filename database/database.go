package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


func GetExec() (*gorm.DB) {
	dsn := "host=db dbname=pg1 user=admin password=REDACTED port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		panic(err)
	}
	return db
}

// func InitDB() {
// 	dsn := "host=localhost dbname=pg2 user=admin password=REDACTED port=22201 sslmode=disable"
// 	db, err := gorm.Open(postgres.Open(dsn))
// 	if err != nil {
// 		panic(err)
// 	}
// 	db.AutoMigrate(&model.ProofState{}, &model.User{}, &model.UserInfo{},&model.ProofType{}, &model.UserProof{})
// 	var proofStates []model.ProofState = []model.ProofState{
// 		{
// 			State: "Not verified",
// 		},
// 		{
// 			State: "Verified",
// 		},
// 	}
// 	db.Create(&proofStates)

// 	var proofTypes []model.ProofType = []model.ProofType{
// 		{
// 			Description	: "Commitment",
// 		},
// 		{
// 			Description	: "CardNumber",
// 		},
// 	}
// 	db.Create(&proofTypes)

// }