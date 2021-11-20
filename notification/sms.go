package notification

import (
	"fmt"
	"os"

	"github.com/kavenegar/kavenegar-go"
)

func SendSMS(message string, receptor []string, pattern string) {
	fmt.Println(message, receptor)
	apiKey := os.Getenv("SMS_API_KEY")
	api := kavenegar.New(apiKey)
	params := &kavenegar.VerifyLookupParam{
	}
	// api.Call.MakeTTS(receptor[0], message, nil)
	// sender := "REDACTED"
	if res, err := api.Verify.Lookup(receptor[0], pattern, message, params); err != nil {
		switch err := err.(type) {
		case *kavenegar.APIError:
			fmt.Println(err.Error(), err.Status)
		case *kavenegar.HTTPError:
			fmt.Println(err.Error())
		default:
			fmt.Println(err.Error())
		}
	} else {
		// for _, r := range res {
			fmt.Println("MessageID 	= ", res.MessageID)
			fmt.Println("Status    	= ", res.Status)
		// }

	}
}