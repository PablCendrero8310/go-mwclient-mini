package mwclientmini

import (
	"context"
	"encoding/json"

	"github.com/PablCendrero8310/go-mwclient-mini/stream"
)

func (c *MediaWikiClient) listen(ctx context.Context, lastID *string, handler func(RecentChange) error) error {
	streamClient := *c.http
	streamClient.Timeout = 0
	if err := stream.Listen(ctx, &streamClient, lastID, func(b []byte) error {
		var event RecentChange
		if err := json.Unmarshal(b, &event); err != nil {
			return err
		}
		if event.Meta.Domain == "canary" {
			return nil
		}
		if event.ServerName != c.Hostname() {
			return nil
		}
		if err := handler(event); err != nil {
			return err
		}
		return nil
	}, c.userAgent); err != nil {
		return err
	}
	return nil
}

func (c *MediaWikiClient) run(ctx context.Context, handler func(RecentChange) error) error {
	streamClient := *c.http
	streamClient.Timeout = 0
	if err := stream.Run(ctx, &streamClient, func(b []byte) error {
		var event RecentChange
		if err := json.Unmarshal(b, &event); err != nil {
			return err
		}
		if event.Meta.Domain == "canary" {
			return nil
		}
		if event.ServerName != c.Hostname() {
			return nil
		}
		if err := handler(event); err != nil {
			return err
		}
		return nil
	}, c.userAgent); err != nil {
		return err
	}
	return nil
}
