package reaction

import "kano/internal/utils/messageutil"

func Main(c *messageutil.MessageContext) error {
	err := VoReactApproval(c)
	if err != nil {
		c.Logger.Errorf("%s", err)
	}

	// Hmm, it returns the last err value, but idc tho
	// Just read the logs
	return err
}
