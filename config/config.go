package config

import (
	"os"
	"encoding/json"
)

type cfg struct {
	DBHost 			string 	`json:"dbHost"`
	DBPort			string	`json:"dbPort"`
	DBName			string	`json:"dbName"`
	DBUseraName		string	`json:"dbUserName"`
	DBPassword		string	`json:"dbPassword"`
	ENVPath 		string 	`json:"envPath"`
	LocalPort		string	`json:"localPort"`
	BaseFilePath	string	`json:"baseFilePath"`

}

var Config cfg

func LoadConfig() error{
	file, err := os.Open("/app/config_server.json")
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&Config)
	if err != nil {
		return err
	}
	return nil
}