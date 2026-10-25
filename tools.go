package mwclientmini

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

func (c *MediaWikiClient) GetCSRF() (_ string, retErr error) {
	var decoded struct {
		Query struct {
			Tokens struct {
				CSRFToken string `json:"csrftoken"`
			} `json:"tokens"`
		} `json:"query"`
	}
	params := url.Values{
		"action": []string{"query"},
		"meta":   []string{"tokens"},
		"format": []string{"json"},
	}
	res, err := c.Get(params)
	if err != nil {
		return "", fmt.Errorf("requesting CSRF token: %w", err)
	}

	if err := json.NewDecoder(res.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	defer func() {
		if err := closeBody(res); err != nil {
			retErr = errors.Join(retErr, err)
		}
	}()
	return decoded.Query.Tokens.CSRFToken, nil
}
