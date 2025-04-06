package message

import (
	"fmt"
	"time"
)

func (ctx MessageContext) PingHandler() {
	msgTime := ctx.Instance.Event.Info.Timestamp.UnixMilli()
	currTime := time.Now().UnixMilli()
	diff := float64((currTime - msgTime)) / 1000

	if diff < 0 {
		diff = -diff
	}

	fmt.Println(msgTime, currTime)

	rep, err := ctx.Instance.Reply(fmt.Sprintf("Pong!\nDiff time: %.3f ± 1 s", diff), true)
	fmt.Printf("Rep: %+v\nErr: %+v", rep, err)
}
