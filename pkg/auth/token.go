package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"groundcover.com/pkg/utils"
)

const (
	TOKEN_ENDPOINT    = "token"
	TOKEN_STORAGE_KEY = "auth.json"
	JWKS_ENDPOINT     = "/.well-known/jwks.json"
)

type Auth0Token struct {
	Claims       Claims `json:"-"`
	ExpiresIn    int64  `json:"expires_in"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type Claims struct {
	jwt.RegisteredClaims
	Scope    string `json:"scope"`
	Org      string `json:"https://client.info/org"`
	Email    string `json:"https://client.info/email"`
	TenantID uint32 `json:"https://client.info/tenant-id"`
}

func (auth0Token *Auth0Token) Load() error {
	var err error

	var data []byte
	if data, err = utils.PresistentStorage.Read(TOKEN_STORAGE_KEY); err != nil {
		return err
	}

	if err = json.Unmarshal(data, &auth0Token); err != nil {
		return err
	}

	if err = auth0Token.loadClaims(); err != nil {
		return err
	}

	return nil
}

func (auth0Token *Auth0Token) Save() error {
	var err error

	var data []byte
	if data, err = json.Marshal(auth0Token); err != nil {
		return err
	}

	utils.PresistentStorage.Write(TOKEN_STORAGE_KEY, data)

	return nil
}

func (auth0Token *Auth0Token) BearerToken() (string, error) {
	var err error

	err = auth0Token.Claims.Valid()

	if errors.Is(err, jwt.ErrTokenExpired) {
		err = auth0Token.RefreshAndSave()
	}

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Bearer %s", auth0Token.AccessToken), nil
}

func (auth0Token *Auth0Token) Fetch(data url.Values) error {
	var err error

	var body []byte
	if body, err = DefaultClient.PostForm(TOKEN_ENDPOINT, data); err != nil {
		return err
	}

	if err = json.Unmarshal(body, &auth0Token); err != nil {
		return err
	}

	if err = auth0Token.loadClaims(); err != nil {
		return err
	}

	return nil
}

func (auth0Token *Auth0Token) RefreshAndSave() error {
	var err error

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("client_id", DefaultClient.ClientId)
	data.Set("refresh_token", auth0Token.RefreshToken)

	var body []byte
	if body, err = DefaultClient.PostForm(TOKEN_ENDPOINT, data); err != nil {
		return err
	}

	if err = json.Unmarshal(body, &auth0Token); err != nil {
		return err
	}

	if err = auth0Token.loadClaims(); err != nil {
		return err
	}

	if err = auth0Token.Claims.Valid(); err != nil {
		return err
	}

	if err = auth0Token.Save(); err != nil {
		return err
	}

	return nil
}

func (auth0Token *Auth0Token) loadClaims() error {
	var err error

	var jwksUrl *url.URL
	if jwksUrl, err = DefaultClient.JoinPath(JWKS_ENDPOINT); err != nil {
		return err
	}

	var jwks *keyfunc.JWKS
	if jwks, err = keyfunc.Get(jwksUrl.String(), keyfunc.Options{}); err != nil {
		return err
	}

	if _, err = jwt.ParseWithClaims(auth0Token.AccessToken, &auth0Token.Claims, jwks.Keyfunc); err != nil {
		return err
	}

	return nil
}
