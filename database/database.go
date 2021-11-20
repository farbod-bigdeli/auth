package database

import (
	"fmt"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"example.com/models/config"

)

var once = sync.Once{}

var g_db *gorm.DB = nil

func createDB() {
	host := config.Config.DBHost
	dbName := config.Config.DBName
	port := config.Config.DBPort
	password := config.Config.DBPassword
	dsn  := fmt.Sprintf("host=%s dbname=%s user=admin password=%s port=%s sslmode=disable", host, dbName, password, port)

	sql := postgres.Open(dsn)
	db, err := gorm.Open(sql)
	if err != nil {
		panic(err)
	}
	g_db = db
}

func GetExec() *gorm.DB {
	once.Do(createDB)

	return g_db
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
