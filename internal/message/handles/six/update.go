package six

import (
	"kano/internal/config"
	"kano/internal/cronjobs"
	"kano/internal/utils/messageutil"
	"os"
	"path"
)

func updateHandler(c *messageutil.MessageContext) error {
	if !c.IsSenderSame(config.GetConfig().OwnerJID) {
		c.QuoteReply("Perintah ini hanya bisa dieksekusi oleh pemilik bot.")
		return nil
	}

	args := c.Parser.Args
	if len(args) > 1 {
		kh := args[1].Content.Data
		os.WriteFile(path.Join("secrets", "khongguan"), []byte(kh), 0644)
	}

	cronjobs.SixUpdateSchedules(c.Client.GetClient())()
	return nil
}
