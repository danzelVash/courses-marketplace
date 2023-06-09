package service

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"io/ioutil"
	courses "mom"
	"os"
	"strings"
	"time"
)

type UrlParams struct {
	clientID     string
	clientSecret string
	redirectURI  string
	scope        []string
	state        string
}

func (a *AuthService) GenerateLinkForOauth() string {
	redUrlParams := UrlParams{
		clientID:    viper.GetString("vk.clientId"),
		redirectURI: viper.GetString("vk.redirectURI"),
		scope:       []string{"account"},
		state:       os.Getenv("VK_STATE"),
	}

	url := fmt.Sprintf("https://oauth.vk.com/authorize?"+
		"&client_id=%s"+
		"&redirect_uri=%s"+
		"&scope=%s"+
		"response_type=code"+
		"&state=%s",
		redUrlParams.clientID, redUrlParams.redirectURI, strings.Join(redUrlParams.scope, "+"), redUrlParams.state)

	return url
}

func (a *AuthService) ValidateParamsFromRedirect(stateTemp, code string) bool {
	if stateTemp == "" || stateTemp != os.Getenv("VK_STATE") || code == "" {
		return false
	}

	return true
}

func (a *AuthService) GenerateUrlForVKHandshake(code string) string {
	handshakeParams := UrlParams{
		clientID:     viper.GetString("vk.clientId"),
		clientSecret: os.Getenv("VK_CLIENT_SECRET"),
		redirectURI:  viper.GetString("vk.redirectURI"),
	}
	url := fmt.Sprintf("https://oauth.vk.com/access_token?"+
		"grant_type=authorization_code"+
		"&code=%s"+
		"&redirect_uri=%s"+
		"&client_id=%s"+
		"&client_secret=%s", code, handshakeParams.redirectURI, handshakeParams.clientID, handshakeParams.clientSecret)

	return url
}

func (a *AuthService) GenerateUrlForVkApi(body io.ReadCloser) (string, error) {
	token := struct {
		AccessToken string `json:"access_token"`
	}{}

	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		a.logger.Errorf("error while reading json from vk response: %s", err.Error())
		return "", err
	}

	err = json.Unmarshal(bytes, &token)
	if err != nil {
		a.logger.Errorf("error while unmarshalling json from vk response: %s", err.Error())
		return "", err
	}

	return fmt.Sprintf("https://api.vk.com/method/%s?v=5.131&access_token=%s", "users.get", token.AccessToken), nil
}

func (a *AuthService) AuthorizeVkUser(body io.ReadCloser) (string, error) {
	vkUser := struct {
		Response []struct {
			LastName  string `json:"last_name"`
			FirstName string `json:"first_name"`
			VkId      int    `json:"id"`
		} `json:"response"`
	}{}

	bytes, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(bytes, &vkUser)
	if err != nil {
		return "", err
	}

	user := courses.User{}
	user.FirstName = vkUser.Response[0].FirstName
	user.LastName = vkUser.Response[0].LastName
	user.Vk = true
	user.VkId = vkUser.Response[0].VkId

	userId, err := a.repos.CheckVkId(user.VkId)
	switch err {
	case courses.ErrNoRows:
		userId, err = a.repos.CreateUser(user)
		if err != nil {
			a.logger.Errorf("error while creating: user with VkId = %d not exists in db : %s", user.VkId, err.Error())
			return "", err
		}
	case nil:
		a.logger.Infof("user with VkId = %d exists in db", user.VkId)
	default:
		a.logger.Errorf("unknown error occured: %s", err.Error())
		return "", err
	}
	SID := generateRandomString(sessLen)
	expiredDate := time.Now().AddDate(0, 0, 15).Round(time.Hour)

	if _, err := a.repos.CreateSession(userId, SID, expiredDate); err != nil {
		a.logger.Errorf("error while creating session for user_id = %d (user was created): %s", user.VkId, err.Error())
		return "", err
	}

	return SID, nil
}
