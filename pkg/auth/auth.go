package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	jwt_modern_claims "github.com/golang-jwt/jwt/v4"
	"github.com/sirupsen/logrus"
)

const (
	AUTH0_APP_CLIENT_ID                    = "UkQmsxoqC8OzajqptiADtAZD6GS2mG9U"
	AUTH0_GET_TOKEN_URL                    = "https://auth.groundcover.com/oauth/token"
	AUTH0_GET_DEVICE_CODE_URL              = "https://auth.groundcover.com/oauth/device/code"
	API_KEY_GENERATE_URL                   = "https://app.groundcover.com/api/system/generate-api-key"
	AUTH_APP_SCOPE                         = "access:router offline_access"
	AUTH0_APP_AUDIENCE                     = "https://groundcover"
	AUTH0_GET_TOKEN_PAYLOAD_TEMPLATE       = "grant_type=urn:ietf:params:oauth:grant-type:device_code&device_code=%s&client_id=%s"
	AUTH0_REFRESH_TOKEN_PAYLOAD_TEMPLATE   = "grant_type=refresh_token&client_id=%s&refresh_token=%s"
	AUTH0_GET_DEVICE_CODE_PAYLOAD_TEMPLATE = "client_id=%s&scope=%s&audience=%s"
	GROUNDCOVER_AUTH_PATH                  = ".groundcover"
	GROUNDCOVER_AUTH_FILE                  = "auth.json"
	GROUNDCOVER_API_KEY_FILE               = "api_key.json"
	AUTH0_MIN_INTERVAL                     = time.Second * 5
)

type CustomClaims struct {
	Scope            string `json:"scope"`
	Org              string `json:"https://client.info/org"`
	Email            string `json:"https://client.info/email"`
	TenantID         uint32 `json:"https://client.info/tenant-id"`
	expectedAudience string `json:"-"`
	auth0Tenant      string `json:"-"`
	jwt_modern_claims.RegisteredClaims
}
type Auth0Error struct {
	Error     string `json:"error"`
	ErrorDesc string `json:"error_description"`
}

type Auth0Token struct {
	Token        string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

type ApiKey struct {
	ApiKey string `json:"apiKey"`
}

type GetApiKeyError struct {
	Message string `json:"message"`
}

type DeviceCodeFlow struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
	VerificationURIComplete string `json:"verification_uri_complete"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
}

// FetchAndSaveApiKey returns whether the user is currently authenticated. This includes whether they have
// existing credentials and whether those are actually valid.
func FetchAndSaveApiKey() (*CustomClaims, error) {
	token, err := MustLoadDefaultCredentials()
	if err != nil {
		return nil, err
	}

	tokenSet, err := refreshTokenSet(token.RefreshToken)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh access token")
	}

	customClaims, err := getCustomClaims(token.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to get custom claims: %s", err.Error())
	}

	authKey, err := getGroundcoverApiKey(tokenSet)
	if err != nil {
		return nil, err
	}

	err = SaveAuthFile(GROUNDCOVER_API_KEY_FILE, authKey)
	if err != nil {
		return nil, err
	}

	return customClaims, nil
}

// MustLoadDefaultCredentials loads the default credentials for the user.
func MustLoadDefaultCredentials() (*Auth0Token, error) {
	token, err := LoadDefaultCredentials()
	if err != nil && os.IsNotExist(err) {
		return nil, errors.New("you must be logged in to perform this operation. Please run `groundcover login`")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get auth credentials: %s", err.Error())
	}

	return token, nil
}

// LoadDefaultCredentials loads the default credentials for the user.
func LoadDefaultCredentials() (*Auth0Token, error) {
	groundcoverAuthFilePath, err := EnsurePersistentFileExists(GROUNDCOVER_AUTH_FILE)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(groundcoverAuthFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token := &Auth0Token{}
	if err := json.NewDecoder(f).Decode(token); err != nil {
		return nil, err
	}

	return token, nil
}

// EnsurePersistentFileExists returns and creates the file path is missing.
func EnsurePersistentFileExists(file string) (string, error) {
	u, err := user.Current()
	if err != nil {
		return "", err
	}

	groundcoverDirPath := filepath.Join(u.HomeDir, GROUNDCOVER_AUTH_PATH)
	if _, err := os.Stat(groundcoverDirPath); os.IsNotExist(err) {
		err := os.Mkdir(groundcoverDirPath, 0744)
		if err != nil {
			return "", err
		}
	}

	filePath := filepath.Join(groundcoverDirPath, file)
	return filePath, nil
}

func LoadApiKey() (*ApiKey, error) {
	groundcoverApiKeyPath, err := EnsurePersistentFileExists(GROUNDCOVER_API_KEY_FILE)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(groundcoverApiKeyPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	apiKey := &ApiKey{}
	if err := json.NewDecoder(f).Decode(apiKey); err != nil {
		return nil, err
	}

	return apiKey, nil
}

func getGroundcoverApiKey(token *Auth0Token) (*ApiKey, error) {
	req, err := http.NewRequest("POST", API_KEY_GENERATE_URL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token.Token))
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, handleGetAPIKeyError(body)
	}

	var apiKey *ApiKey
	err = json.Unmarshal(body, &apiKey)
	if err != nil {
		return nil, err
	}

	return apiKey, nil
}

func handleGetAPIKeyError(body []byte) error {
	var apiKeyError *GetApiKeyError
	err := json.Unmarshal(body, &apiKeyError)
	if err != nil {
		return fmt.Errorf("failed to decode GetAPIKey error response: %s", err.Error())
	}

	return fmt.Errorf("%s", apiKeyError.Message)
}

func getToken(deviceCode, clientID string) (*Auth0Token, error) {
	payload := strings.NewReader(fmt.Sprintf(AUTH0_GET_TOKEN_PAYLOAD_TEMPLATE, deviceCode, clientID))

	req, err := http.NewRequest("POST", AUTH0_GET_TOKEN_URL, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, handleAuth0Error(body)
	}

	var auth *Auth0Token
	err = json.Unmarshal(body, &auth)
	if err != nil {
		return nil, fmt.Errorf("failed to decode auth response: %s", err.Error())
	}

	return auth, nil
}

func handleAuth0Error(body []byte) error {
	var auth0Error *Auth0Error
	err := json.Unmarshal(body, &auth0Error)
	if err != nil {
		return fmt.Errorf("failed to decode Auth0 error response: %s", err.Error())
	}

	return fmt.Errorf("%s: %s", auth0Error.Error, auth0Error.ErrorDesc)
}

func getDeviceCodeFlow() (*DeviceCodeFlow, error) {
	payload := strings.NewReader(fmt.Sprintf(AUTH0_GET_DEVICE_CODE_PAYLOAD_TEMPLATE, AUTH0_APP_CLIENT_ID, AUTH_APP_SCOPE, AUTH0_APP_AUDIENCE))

	req, err := http.NewRequest("POST", AUTH0_GET_DEVICE_CODE_URL, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var deviceCodeFlow *DeviceCodeFlow
	err = json.Unmarshal(body, &deviceCodeFlow)
	if err != nil {
		return nil, err
	}

	return deviceCodeFlow, nil
}

func SaveAuthFile(fname string, apiKey interface{}) error {
	path, err := EnsurePersistentFileExists(fname)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(apiKey); err != nil {
		return err
	}

	return nil
}

func Login(ctx context.Context, manualMode bool) error {
	deviceCodeFlow, err := getDeviceCodeFlow()
	if err != nil {
		return fmt.Errorf("failed to get device code flow: %s", err.Error())
	}

	fmt.Printf("Device confirmation code, make sure you see it in your browser: '%s'\n", deviceCodeFlow.UserCode)

	if manualMode {
		fmt.Printf("In order to login, browse: %s\n", deviceCodeFlow.VerificationURIComplete)
	} else {
		cmd := exec.Command("xdg-open", deviceCodeFlow.VerificationURIComplete)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("failed to open browser, try running with --manual flag")
		}

		fmt.Printf("In order to login, browse: %s\n", deviceCodeFlow.VerificationURIComplete)
	}

	token, err := pollToken(deviceCodeFlow, ctx)
	if err != nil {
		return err
	}

	if err = SaveAuthFile(GROUNDCOVER_AUTH_FILE, token); err != nil {
		return fmt.Errorf("failed to persist auth token: %s", err.Error())
	}

	fmt.Print("You are successfully logged in!\n")
	return nil
}

func pollToken(deviceCodeFlow *DeviceCodeFlow, ctx context.Context) (*Auth0Token, error) {
	ticker := time.NewTicker(AUTH0_MIN_INTERVAL)

	for {
		select {
		case <-ticker.C:
			token, err := getToken(deviceCodeFlow.DeviceCode, AUTH0_APP_CLIENT_ID)
			if err != nil {
				logrus.Debugf("Failed to poll token: %s", err.Error())
				continue
			}
			return token, nil
		case <-ctx.Done():
			return nil, errors.New("timed out while waiting for login")
		}
	}
}

func refreshTokenSet(refreshToken string) (*Auth0Token, error) {
	payload := strings.NewReader(fmt.Sprintf(AUTH0_REFRESH_TOKEN_PAYLOAD_TEMPLATE, AUTH0_APP_CLIENT_ID, refreshToken))

	req, err := http.NewRequest("POST", AUTH0_GET_TOKEN_URL, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Add("content-type", "application/x-www-form-urlencoded")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("invalid refresh token")
	}

	var auth *Auth0Token
	err = json.Unmarshal(body, &auth)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token refresh response: %s", err.Error())
	}

	if err = SaveAuthFile(GROUNDCOVER_AUTH_FILE, auth); err != nil {
		return nil, fmt.Errorf("failed to persist auth token: %s", err.Error())
	}

	return auth, nil
}

func getCustomClaims(tokenString string) (*CustomClaims, error) {
	parsedToken, err := jwt.Parse(tokenString, nil)
	if parsedToken == nil {
		return nil, err
	}

	var customClaims *CustomClaims
	claimsBytes, err := json.Marshal(parsedToken.Claims)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(claimsBytes, &customClaims)
	if err != nil {
		return nil, err
	}

	return customClaims, nil
}
