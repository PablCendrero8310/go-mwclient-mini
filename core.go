// Package gomwclientmini Go mini wrapper for MediaWiki API
package gomwclientmini

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// A MediaWikiClient abstract the MediaWiki API
type MediaWikiClient struct {
	APIURL    string
	USERAGENT string
	HTTP      *http.Client
	USER      *url.Userinfo
}

// NewMediaWikiClient is the constructor of MediaWikiClient
func NewMediaWikiClient(apiURL, username, password, userAgent string) (*MediaWikiClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creando cookie jar: %w", err)
	}

	return &MediaWikiClient{
		APIURL:    apiURL,
		USERAGENT: userAgent,
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
		USER: url.UserPassword(username, password),
	}, nil
}

// Login allows to log in the API using a Bot Password
func (c *MediaWikiClient) Login() error {
	// TODO: allow login with OAuth
	params := url.Values{
		"action": {"query"},
		"meta":   {"tokens"},
		"type":   {"login"},
		"format": {"json"},
	}

	req, err := http.NewRequest(
		http.MethodGet,
		c.APIURL+"?"+params.Encode(),
		nil,
	)
	if err != nil {
		return fmt.Errorf("creando petición de token: %w", err)
	}

	req.Header.Set("User-Agent", c.USERAGENT)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("solicitando token de inicio de sesión: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		status := resp.Status
		resp.Body.Close()
		return fmt.Errorf("solicitando token: HTTP %s", status)
	}

	var tokenData struct {
		Query struct {
			Tokens struct {
				LoginToken string `json:"logintoken"`
			} `json:"tokens"`
		} `json:"query"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenData)
	resp.Body.Close()

	if err != nil {
		return fmt.Errorf("decodificando token de inicio de sesión: %w", err)
	}

	token := tokenData.Query.Tokens.LoginToken
	if token == "" {
		return fmt.Errorf("MediaWiki no devolvió el token de inicio de sesión")
	}

	password, _ := c.USER.Password()

	form := url.Values{
		"action":     {"login"},
		"lgname":     {c.USER.Username()},
		"lgpassword": {password},
		"lgtoken":    {token},
		"format":     {"json"},
	}

	req, err = http.NewRequest(
		http.MethodPost,
		c.APIURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("creando petición de inicio de sesión: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.USERAGENT)

	resp, err = c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("enviando inicio de sesión: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("inicio de sesión: HTTP %s", resp.Status)
	}

	var loginData struct {
		Login struct {
			Result string `json:"result"`
			Reason string `json:"reason"`
		} `json:"login"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&loginData); err != nil {
		return fmt.Errorf("decodificando respuesta de inicio de sesión: %w", err)
	}

	if loginData.Login.Result != "Success" {
		return fmt.Errorf(
			"inicio de sesión fallido: %s (%s)",
			loginData.Login.Result,
			loginData.Login.Reason,
		)
	}

	return nil
}

func (c *MediaWikiClient) Get(params url.Values) (*http.Response, error) {
	// No modificar los parámetros del llamador.
	query := make(url.Values, len(params))
	for key, values := range params {
		query[key] = append([]string(nil), values...)
	}
	query.Set("format", "json")
	query.Set("formatversion", "2")

	req, err := http.NewRequest(
		http.MethodGet,
		c.APIURL+"?"+query.Encode(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("creando petición a MediaWiki: %w", err)
	}

	req.Header.Set("User-Agent", c.USERAGENT)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("realizando petición a MediaWiki: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		status := resp.Status
		resp.Body.Close()
		return nil, fmt.Errorf("petición a MediaWiki: HTTP %s", status)
	}

	return resp, nil
}
