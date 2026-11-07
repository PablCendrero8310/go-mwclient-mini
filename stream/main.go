package stream

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const streamURL = "https://stream.wikimedia.org/v2/stream/recentchange"

func consume(ctx context.Context, client *http.Client, lastID *string, handler func([]byte), userAgent string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("User-Agent", userAgent)

	if *lastID != "" {
		req.Header.Set("Last-Event-ID", *lastID)
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status: %s", resp.Status)
	}

	reader := bufio.NewScanner(resp.Body)
	reader.Buffer(make([]byte, 4096), 1024*1024)

	var data []string
	var eventID string

	dispatch := func() {
		if len(data) == 0 {
			return
		}

		handler([]byte(strings.Join(data, "\n")))

		if eventID != "" {
			*lastID = eventID
		}

		data = nil
	}

	for reader.Scan() {
		line := reader.Text()

		if line == "" {
			dispatch()
			eventID = ""
			continue
		}

		if strings.HasPrefix(line, ":") {
			continue // Comentario / heartbeat
		}

		field, value, found := strings.Cut(line, ":")
		if !found {
			value = ""
		} else {
			value = strings.TrimPrefix(value, " ")
		}

		switch field {
		case "data":
			data = append(data, value)
		case "id":
			eventID = value
		}
	}

	if err := reader.Err(); err != nil {
		return err
	}

	return io.EOF
}
