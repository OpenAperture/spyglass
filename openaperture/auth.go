package openaperture

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path"
)

//Auth struct
type Auth struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   string `json:"expires_in"`
	Scope       string `json:"scope"`
}

//GetAuthorizationHeader build a proper authorization header for all api calls
func (oauth *Auth) GetAuthorizationHeader() string {
	return "Bearer access_token=" + oauth.AccessToken
}

// GetAuth returns authentication config from various sources
func GetAuth() (*Auth, error) {
	if _, err := os.Stat(path.Join(os.Getenv("HOME"), ".aperturecfg")); os.IsNotExist(err) {
		return EnvAuth()
	}
	return SharedAuth()
}

// SharedAuth generates an authentication config from local json file
func SharedAuth() (*Auth, error) {
	var config map[string]string
	configPath := path.Join(os.Getenv("HOME"), ".aperturecfg")
	configBytes, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(configBytes, &config)
	return NewAuth(config["username"], config["password"])
}

// EnvAuth pulls authentication information from environment variables
func EnvAuth() (*Auth, error) {
	username := os.Getenv("APERTURE_USERNAME")
	password := os.Getenv("APERTURE_PASSWORD")
	if username == "" || password == "" {
		return nil, errors.New("username or password is blank")
	}
	return NewAuth(username, password)
}

// NewAuth requests a new token from idp
func NewAuth(username string, password string) (*Auth, error) {
	var auth Auth
	url := "https://auth.psft.co/oauth/token"
	credentials := map[string]string{
		"grant_type": "password",
		"username":   username,
		"password":   password,
	}
	payload, _ := json.Marshal(credentials)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &auth)
	return &auth, nil
}
