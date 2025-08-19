package cronfinaid

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	finaiditb "kano/internals/utils/finaid-itb"
	"kano/internals/utils/kanoutils"
	"os"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

var (
	ErrNoData            = errors.New("server returned empty array")
	ErrNoJIDs            = errors.New("no registered JIDs")
	ErrAlreadyRegistered = errors.New("already registered")
	ErrNotRegistered     = errors.New("not registered")
)

type Config struct {
	LastID         int             `json:"last_id"`
	RegisteredJIDs map[string]bool `json:"registered_jids"`
}

var defaultConfig = Config{
	LastID:         0,
	RegisteredJIDs: map[string]bool{},
}

func getConfig() Config {
	fileConf, err := os.ReadFile("local/finaid.json")
	if err != nil {
		return defaultConfig
	}

	var conf Config
	err = json.Unmarshal(fileConf, &conf)
	if err != nil {
		return defaultConfig
	}

	return conf
}

func saveConfig(c Config) error {
	mar, err := json.Marshal(c)
	if err != nil {
		return err
	}

	err = os.WriteFile("local/finaid.json", mar, 0666)
	if err != nil {
		return err
	}

	return nil
}

func send(cli *whatsmeow.Client, jid, msg string) {
	parsed, err := types.ParseJID(jid)
	if err != nil {
		return
	}

	cli.SendMessage(context.Background(), parsed, &waE2E.Message{
		Conversation: &msg,
	})
}

func fetchAndSend(cli *whatsmeow.Client) error {
	fmt.Println("[FINAID SCHOLARSHIP] Retrieving config")
	conf := getConfig()
	if len(conf.RegisteredJIDs) == 0 {
		return ErrNoJIDs
	}

	fmt.Println("[FINAID SCHOLARSHIP] Retrieving scholarships")
	sc, err := finaiditb.FetchScholarships(5)
	if err != nil {
		return err
	}

	fmt.Printf("[FINAID SCHOLARSHIP] Got data length %d\n", len(sc.Data))
	if len(sc.Data) == 0 {
		return ErrNoData
	}

	fmt.Println("[FINAID SCHOLARSHIP] Check if LastID is equal to the first of data ID")
	if conf.LastID == sc.Data[0].ID {
		return nil
	}

	fmt.Println("[FINAID SCHOLARSHIP] Loop through array of data")
	for _, data := range sc.Data {
		if data.ID == conf.LastID {
			fmt.Println("[FINAID SCHOLARSHIP] data.ID is equals to conf.LastID, break")
			break
		}

		fmt.Println("[FINAID SCHOLARSHIP] Generating message")
		msg := kanoutils.GenerateFinaidScholarshipMessage(data)
		for jid := range conf.RegisteredJIDs {
			fmt.Printf("[FINAID SCHOLARSHIP] Sending to %s\n", jid)
			send(cli, jid, msg)
		}
	}
	conf.LastID = sc.Data[0].ID
	fmt.Println("[FINAID SCHOLARSHIP] End")

	fmt.Println("[FINAID SCHOLARSHIP] Saving config")
	return saveConfig(conf)
}

func FinaidScholarshipCronFunc(client *whatsmeow.Client) {
	for range 5 {
		fmt.Println("[FINAID SCHOLARSHIP] Start")
		err := fetchAndSend(client)
		if err == nil {
			break
		}

		fmt.Println("[FINAID SCHOLARSHIP] Failed with error:", err)
	}
}

func RegisterNewJID(jid types.JID) error {
	c := getConfig()
	_, ok := c.RegisteredJIDs[jid.String()]
	if ok {
		return ErrAlreadyRegistered
	}

	c.RegisteredJIDs[jid.String()] = true

	return saveConfig(c)
}

func UnregisterJID(jid types.JID) error {
	c := getConfig()
	_, ok := c.RegisteredJIDs[jid.String()]
	if !ok {
		return ErrNotRegistered
	}

	delete(c.RegisteredJIDs, jid.String())

	return saveConfig(c)
}
