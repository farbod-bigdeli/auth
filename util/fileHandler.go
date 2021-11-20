package util

import (
	"fmt"
	"io"
	"os"
	"strings"

	"example.com/models/config"
	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)


func UploadPhoto(c gin.Context, path string) (string, error) {
	BasePath := config.Config.BaseFilePath

	file, header, err := c.Request.FormFile("photo")
	fmt.Println("\n",header.Size)
	mimeArr := strings.Split(header.Header["Content-Type"][0], "/")
	mime := mimeArr[1]
	if err != nil {
		return "", err
	}

	if mime != "jpeg" && mime != "png" {
		return "", fmt.Errorf("WrongType")
	}
	if header.Size > 2000000 {
		return "", fmt.Errorf("SizeExceed")

	}
	
    if _, err = os.Stat(BasePath + path); os.IsNotExist(err) {
		os.Mkdir(BasePath + path, os.ModePerm)
	}
	filePath := BasePath + path + uuid.NewV4().String() + "." + mime
	fmt.Println(filePath)
	out, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
		panic(err)
	}
	return filePath, nil

}