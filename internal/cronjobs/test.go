package cronjobs

import (
	"fmt"

	"go.mau.fi/whatsmeow"
)

func TestCronJob(_cli *whatsmeow.Client) func() {
	return func() {
		fmt.Println("TestCronJob called")
	}
}
