package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/spf13/viper"
)

type GoogleOAuthToken struct {
	AccessToken string
	IDToken     string
}

type GoogleUserResult struct {
	Id            string
	Email         string
	VerifiedEmail bool
	Name          string
	GivenName     string
	FamilyName    string
	Picture       string
	Locale        string
}

func GetGoogleOAuthToken(code string) (*GoogleOAuthToken, error) {
	const rootURl = "https://oauth2.googleapis.com/token"

	values := url.Values{}
	values.Add("grant_type", "authorization_code")
	values.Add("code", code)
	values.Add("client_id", viper.GetString("google_oauth_client_id"))
	values.Add("client_secret", viper.GetString("google_oauth_client_secret"))
	values.Add("redirect_uri", viper.GetString("google_oauth_redirect_url"))

	query := values.Encode()

	req, err := http.NewRequest("POST", rootURl, bytes.NewBufferString(query))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	client := http.Client{
		Timeout: time.Second * 30,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("could not retrieve token")
	}

	var resBody bytes.Buffer
	_, err = io.Copy(&resBody, res.Body)
	if err != nil {
		return nil, err
	}

	var GoogleOAuthTokenRes map[string]interface{}

	if err := json.Unmarshal(resBody.Bytes(), &GoogleOAuthTokenRes); err != nil {
		return nil, err
	}

	tokenBody := &GoogleOAuthToken{
		AccessToken: GoogleOAuthTokenRes["access_token"].(string),
		IDToken:     GoogleOAuthTokenRes["id_token"].(string),
	}

	return tokenBody, nil
}

func GetGoogleUser(accessToken string, idToken string) (*GoogleUserResult, error) {
	rootUrl := fmt.Sprintf("https://www.googleapis.com/oauth2/v1/userinfo?alt=json&access_token=%s", accessToken)

	req, err := http.NewRequest("GET", rootUrl, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", idToken))

	client := http.Client{
		Timeout: time.Second * 30,
	}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("could not retrieve user")
	}

	var resBody bytes.Buffer
	_, err = io.Copy(&resBody, res.Body)
	if err != nil {
		return nil, err
	}

	var GoogleUserRes map[string]interface{}

	if err := json.Unmarshal(resBody.Bytes(), &GoogleUserRes); err != nil {
		return nil, err
	}

	userBody := &GoogleUserResult{
		Id:            GoogleUserRes["id"].(string),
		Email:         GoogleUserRes["email"].(string),
		VerifiedEmail: GoogleUserRes["verified_email"].(bool),
		Name:          GoogleUserRes["name"].(string),
		GivenName:     GoogleUserRes["given_name"].(string),
		Picture:       GoogleUserRes["picture"].(string),
		Locale:        GoogleUserRes["locale"].(string),
	}

	return userBody, nil
}
