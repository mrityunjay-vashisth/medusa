package oauth

import (
	"golang.org/x/oauth2"
)

type oAuthProvider interface {
	GetConfig() *oauth2.Config
	GetUserInfo(token *oauth2.Token) (map[string]interface{}, error)
}
