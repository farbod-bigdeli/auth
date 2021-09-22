package database

import (
	"example.com/models/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)


func GetExec() (*gorm.DB) {
	dsn := "host=localhost dbname=pg2 user=admin password=REDACTED port=22201 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		panic(err)
	}
	return db
}

func InitDB() {
	dsn := "host=localhost dbname=pg2 user=admin password=REDACTED port=22201 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn))
	if err != nil {
		panic(err)
	}
	db.AutoMigrate(&model.ProofState{}, &model.User{}, &model.UserInfo{},&model.ProofType{}, &model.UserProof{})

}