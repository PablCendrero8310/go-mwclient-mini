package mwclientmini

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
	apiURL    string
	userAgent string
	http      *http.Client
	user      *url.Userinfo
}

// UserAgent getter for userAgent
func (c *MediaWikiClient) UserAgent() string {
	return c.userAgent
}

// SetUserAgent setter for userAgent
func (c *MediaWikiClient) SetUserAgent(userAgent string) {
	c.userAgent = userAgent
}

// User setter for user
func (c *MediaWikiClient) User() url.Userinfo {
	return *c.user
}

// SetUserPassword setter for user
func (c *MediaWikiClient) SetUserPassword(username, password string) {
	c.user = url.UserPassword(username, password)
}

// New the constructor of MediaWikiClient
func New(apiURL, username, password, userAgent string) (*MediaWikiClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("creating cookie jar: %w", err)
	}

	return &MediaWikiClient{
		apiURL:    apiURL,
		userAgent: userAgent,
		http: &http.Client{
			Timeout: 30 * time.Second,
			Jar:     jar,
		},
		user: url.UserPassword(username, password),
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
		c.apiURL+"?"+params.Encode(),
		nil,
	)
	if err != nil {
		return fmt.Errorf("requesting token: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("requesting token: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		status := resp.Status
		resp.Body.Close()
		return fmt.Errorf("requesting token: HTTP %s", status)
	}

	var tokenData struct {
		Query struct {
			Tokens struct {
				LoginToken string `json:"logintoken"`
			} `json:"tokens"`
		} `json:"query"`
	}

	err = json.NewDecoder(resp.Body).Decode(&tokenData)
	if err != nil {
		return fmt.Errorf("decoding login token: %w", err)
	}

	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("closing respose: %w", err)
	}
	token := tokenData.Query.Tokens.LoginToken
	if token == "" {
		return fmt.Errorf("MediaWiki didn't return the login token")
	}

	password, _ := c.user.Password()

	form := url.Values{
		"action":     {"login"},
		"lgname":     {c.user.Username()},
		"lgpassword": {password},
		"lgtoken":    {token},
		"format":     {"json"},
	}

	req, err = http.NewRequest(
		http.MethodPost,
		c.apiURL,
		strings.NewReader(form.Encode()),
	)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.userAgent)

	resp, err = c.http.Do(req)
	if err != nil {
		return fmt.Errorf("login: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login: HTTP %s", resp.Status)
	}

	var loginData struct {
		Login struct {
			Result string `json:"result"`
			Reason string `json:"reason"`
		} `json:"login"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&loginData); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	if loginData.Login.Result != "Success" {
		return fmt.Errorf(
			"login failed: %s (%s)",
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
		c.apiURL+"?"+query.Encode(),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("requesting MediaWiki: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting MediaWiki: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		status := resp.Status
		resp.Body.Close()
		return nil, fmt.Errorf("requesting MediaWiki: HTTP %s", status)
	}

	return resp, nil
}

func (c *MediaWikiClient) Post(data url.Values) (*http.Response, error) {
	req, err := http.NewRequest(
		http.MethodPost,
		c.apiURL,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return nil, fmt.Errorf("requesting MediaWiki: %w", err)
	}

	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting MediaWiki: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		status := resp.Status
		resp.Body.Close()
		return nil, fmt.Errorf("requesting MediaWiki: HTTP %s", status)
	}

	return resp, nil
}
