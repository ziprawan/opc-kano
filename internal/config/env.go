package config

import (
	"encoding/base64"
	"fmt"
	"os"

	"go.mau.fi/whatsmeow/types"
)

type Config struct {
	DatabaseURL string
	OwnerJID    types.JID
	PddiktiKey  []byte
	PddiktiIv   []byte
}

func InitConfig() *Config {
	conf := Config{}

	databaseUrl, err := getEnv("DATABASE_URL")
	if err != nil {
		panic(err)
	}
	conf.DatabaseURL = databaseUrl

	if ownerJid, err := getEnv("OWNER_JID"); err == nil {
		conf.OwnerJID, err = types.ParseJID(ownerJid)
		if err != nil {
			panic(err)
		}
	}

	if pddiktiKey, err := getEnv("PDDIKTI_KEY"); err == nil {
		if pddiktiIv, err := getEnv("PDDIKTI_IV"); err == nil {
			keyByte, kerr := base64.StdEncoding.DecodeString(pddiktiKey)
			ivByte, ierr := base64.StdEncoding.DecodeString(pddiktiIv)
			if kerr == nil && ierr == nil {
				conf.PddiktiKey = keyByte
				conf.PddiktiIv = ivByte
			}
		}
	}

	return &conf
}

func getEnv(key string) (string, error) {
	if value, exists := os.LookupEnv(key); exists {
		return value, nil
	}

	return "", fmt.Errorf("cannot find %s in your environment", key)
}
