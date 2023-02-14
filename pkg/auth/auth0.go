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
	ExpiresIn    int64  `json:"expires_in" validate:"required"`
	AccessToken  string `json:"access_token" validate:"required"`
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type Claims struct {
	jwt.RegisteredClaims
	Scope string `json:"scope" validate:"required"`
	Org   string `json:"https://client.info/org" validate:"required"`
	Email string `json:"https://client.info/email" validate:"required"`
}

func LoadAuth0Token() (*Auth0Token, error) {
	var err error

	var data []byte
	if data, err = utils.PresistentStorage.Read(TOKEN_STORAGE_KEY); err != nil {
		return nil, err
	}

	auth0Token := &Auth0Token{}
	err = auth0Token.parseBody(data)

	if errors.Is(err, jwt.ErrTokenExpired) {
		err = auth0Token.RefreshAndSave()
	}

	if err != nil {
		return nil, err
	}

	return auth0Token, nil
}

func (auth0Token *Auth0Token) Save() error {
	var err error

	var data []byte
	if data, err = json.Marshal(auth0Token); err != nil {
		return err
	}

	return utils.PresistentStorage.Write(TOKEN_STORAGE_KEY, data)
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

	return auth0Token.parseBody(body)
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

	if err = auth0Token.parseBody(body); err != nil {
		return err
	}

	return auth0Token.Save()
}

func (auth0Token *Auth0Token) parseBody(body []byte) error {
	var err error

	if err = json.Unmarshal(body, auth0Token); err != nil {
		return err
	}

	if err = auth0Token.loadClaims(); err != nil {
		return err
	}

	if err = validate.Struct(auth0Token); err != nil {
		return err
	}

	return nil
}

func (auth0Token Auth0Token) GetId() string {
	return ""
}

func (auth0Token Auth0Token) GetOrg() string {
	return auth0Token.Claims.Org
}

func (auth0Token Auth0Token) GetEmail() string {
	return auth0Token.Claims.Email
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
