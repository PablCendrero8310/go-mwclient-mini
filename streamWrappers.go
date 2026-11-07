package mwclientmini

import (
	"context"
	"encoding/json"

	"github.com/PablCendrero8310/go-mwclient-mini/stream"
)

func (c *MediaWikiClient) Listen(ctx context.Context, lastID *string, handler func(RecentChange)) error {
	stream.Listen(ctx, c.http, lastID, func(b []byte) {
		var event RecentChange
		if err := json.Unmarshal(b, &event); err != nil {
			return
		}
		if event.Meta.Domain == "canary" {
			return
		}
		if event.ServerName == c.Hostname() {
			return
		}
		handler(event)
	}, c.userAgent)
	return nil
}
