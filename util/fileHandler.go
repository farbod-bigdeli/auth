package util

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

const BasePath = "/app/storage/"

func UploadPhoto(c gin.Context, path string) (string, error) {
	file, header, err := c.Request.FormFile("photo")
	mimeArr := strings.Split(header.Header["Content-Type"][0], "/")
	mime := mimeArr[1]
	if err != nil {
		return "", err
	}

	if mime != "jpeg" && mime != "png" {
		return "", fmt.Errorf("wrong type")
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