package reaction

import (
	"kano/internal/database"
	"kano/internal/utils/messageutil"
)

var db = database.GetInstance()

func Main(c *messageutil.MessageContext) error {
	err := VoReactApproval(c)
	if err != nil {
		c.Logger.Errorf("%s", err)
	}

	err = SawitAcceptChallenge(c)
	if err != nil {
		c.Logger.Errorf("%s", err)
	}

	// Hmm, it returns the last err value, but idc tho
	// Just read the logs
	return err
}
