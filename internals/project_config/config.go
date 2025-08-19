package projectconfig

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"go.mau.fi/whatsmeow/types"
)

type NullString struct {
	String string
	Valid  bool
}

type Config struct {
	SessionName string
	DatabaseURL string
	OwnerJID    types.JID
	JWTSecret   []byte

	// Google OAuth 2.0
	GoogleID     string
	GoogleSecret string

	// These fields are optional

	PDDiktiKey NullString
	PDDiktiIV  NullString
}

var configInstance *Config
var once sync.Once

func LoadConfig() *Config {
	once.Do(func() {
		sessionName, err := getEnv("SESSION_NAME")
		if err != nil {
			panic(err)
		}

		databaseURL, err := getEnv("DATABASE_URL")
		if err != nil {
			panic(err)
		}

		ownerJID, err := getEnv("OWNER_JID")
		if err != nil {
			panic(err)
		}

		JWTSecret, err := getEnv("JWT_SECRET")
		if err != nil {
			panic(err)
		}

		if _, err := strconv.Atoi(ownerJID); err != nil {
			panic(err)
		}

		googleId, err := getEnv("GOOGLE_ID")
		if err != nil {
			panic(err)
		}
		googleSecret, err := getEnv("GOOGLE_SECRET")
		if err != nil {
			panic(err)
		}

		configInstance = &Config{
			SessionName:  sessionName,
			DatabaseURL:  databaseURL,
			OwnerJID:     types.NewJID(ownerJID, "s.whatsapp.net"),
			JWTSecret:    []byte(JWTSecret),
			GoogleID:     googleId,
			GoogleSecret: googleSecret,
		}

		pddiktiKey, _ := getEnv("PDDIKTI_KEY")
		if pddiktiKey != "" {
			configInstance.PDDiktiKey.String = pddiktiKey
			configInstance.PDDiktiKey.Valid = true
		}
		pddiktiIV, _ := getEnv("PDDIKTI_IV")
		if pddiktiIV != "" {
			configInstance.PDDiktiIV.String = pddiktiIV
			configInstance.PDDiktiIV.Valid = true
		}
	})

	return configInstance
}

func GetConfig() *Config {
	if configInstance == nil {
		panic(errors.New("config is not initialized! Call LoadConfig() first"))
	}
	return configInstance
}

func getEnv(key string) (string, error) {
	if value, exists := os.LookupEnv(key); exists {
		return value, nil
	}

	return "", fmt.Errorf("cannot find %s in your environment", key)
}
