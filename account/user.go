package account

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

type User struct {
	Username                   string `json:"username"`
	DisplayName                string `json:"displayName"`
	IsStoryTeller              bool   `json:"isStoryTeller"`
	ExternalCharacterSheetLink string `json:"externalCharacterSheetLink"`
}

func LoadUser(username string) (User, error) {
	var user User

	userInfoBytes, err := os.ReadFile("." + string(filepath.Separator) + filepath.Join("data", "users", username, "info.json"))
	if err != nil {
		err := fmt.Errorf("user file could not be read: %s", err)
		log.Println(err)

		return user, err
	}

	if err := json.Unmarshal(userInfoBytes, &user); err != nil {
		err := fmt.Errorf("user file could not be parsed: %s", err)
		log.Println(err)

		return user, err
	}

	return user, nil
}

func (self User) Authenticate(secret string) bool {
	secretBytes, err := os.ReadFile("." + string(filepath.Separator) + filepath.Join("data", "users", self.Username, "secret.txt"))
	if err != nil {
		log.Printf("secret file could not be read: %s", err)
		return false
	}

	if secret != string(secretBytes) {
		log.Printf("wrong secret for user: %s", self.Username)
		return false
	}

	return true
}
