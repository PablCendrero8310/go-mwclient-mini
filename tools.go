package mwclientmini

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/PablCendrero8310/go-mwclient-mini/options"
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

func (c *MediaWikiClient) EditByName(text string, options options.EditOptions) error {
	var mwErr struct {
		Error *struct {
			Code string `json:"code"`
			Info string `json:"info"`
		} `json:"error"`
	}
	token, err := c.GetCSRF()
	if err != nil {
		return err
	}
	params := options.ToParams()
	params.Set("action", "edit")
	params.Set("text", text)
	params.Set("token", token)
	res, err := c.Post(params)
	if err != nil {
		return err
	}
	if err := json.NewDecoder(res.Body).Decode(&mwErr); err != nil {
		return err
	}

	if mwErr.Error != nil {
		return fmt.Errorf("%s: %s", mwErr.Error.Code, mwErr.Error.Info)
	}
	return nil
}
