package mwclientmini

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

// MediaWikiClient abstracts the MediaWiki API.
type MediaWikiClient struct {
	apiURL    string
	userAgent string
	http      *http.Client
	user      *url.Userinfo
}

// UserAgent returns the configured user agent.
func (c *MediaWikiClient) UserAgent() string {
	return c.userAgent
}

// SetUserAgent sets the user agent.
func (c *MediaWikiClient) SetUserAgent(userAgent string) {
	c.userAgent = userAgent
}

// User returns the configured user information.
func (c *MediaWikiClient) User() url.Userinfo {
	return *c.user
}

// SetUserPassword sets the username and password.
func (c *MediaWikiClient) SetUserPassword(username, password string) {
	c.user = url.UserPassword(username, password)
}

// New constructs a MediaWikiClient.
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

// closeBody closes the response body and wraps any error.
func closeBody(resp *http.Response) error {
	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("closing response body: %w", err)
	}
	return nil
}

// Login logs in to the MediaWiki API using a Bot Password.
func (c *MediaWikiClient) Login() (retErr error) {
	// TODO: Allow login with OAuth.
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
		statusErr := fmt.Errorf("requesting token: HTTP %s", resp.Status)
		return errors.Join(statusErr, closeBody(resp))
	}

	var tokenData struct {
		Query struct {
			Tokens struct {
				LoginToken string `json:"logintoken"`
			} `json:"tokens"`
		} `json:"query"`
	}

	decodeErr := json.NewDecoder(resp.Body).Decode(&tokenData)
	closeErr := closeBody(resp)

	if decodeErr != nil || closeErr != nil {
		return errors.Join(
			wrapError("decoding login token", decodeErr),
			closeErr,
		)
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

	// Propagate a close error even if Login returns another error.
	defer func() {
		if err := closeBody(resp); err != nil {
			retErr = errors.Join(retErr, err)
		}
	}()

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

// Get sends a GET request to the MediaWiki API.
// The caller is responsible for closing the returned response body.
func (c *MediaWikiClient) Get(params url.Values) (*http.Response, error) {
	// Do not modify the caller's parameters.
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
		statusErr := fmt.Errorf("requesting MediaWiki: HTTP %s", resp.Status)
		return nil, errors.Join(statusErr, closeBody(resp))
	}

	return resp, nil
}

// Post sends a POST request to the MediaWiki API.
// The caller is responsible for closing the returned response body.
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
		statusErr := fmt.Errorf("requesting MediaWiki: HTTP %s", resp.Status)
		return nil, errors.Join(statusErr, closeBody(resp))
	}

	return resp, nil
}

// wrapError wraps err with a message, returning nil when err is nil.
func wrapError(message string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}
