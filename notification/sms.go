package notification

import (
	"fmt"
	"os"

	"github.com/kavenegar/kavenegar-go"
)

func SendSMS(message string, receptor []string) {
	fmt.Println(message, receptor)
	apiKey := os.Getenv("SMS_API_KEY")
	api := kavenegar.New(apiKey)
	sender := "REDACTED"                 
	if res, err := api.Message.Send(sender, receptor, message, nil); err != nil {
		switch err := err.(type) {
		case *kavenegar.APIError:
			fmt.Println(err.Error(), err.Status)
		case *kavenegar.HTTPError:
			fmt.Println(err.Error())
		default:
			fmt.Println(err.Error())
		}
	} else {
		for _, r := range res {
			fmt.Println("MessageID 	= ", r.MessageID)
			fmt.Println("Status    	= ", r.Status)
			//...
		}
	}
}