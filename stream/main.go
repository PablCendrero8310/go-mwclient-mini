package stream

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const streamURL = "https://stream.wikimedia.org/v2/stream/recentchange"

func closeBody(resp *http.Response) error {
	if err := resp.Body.Close(); err != nil {
		return fmt.Errorf("closing response body: %w", err)
	}
	return nil
}

func consume(ctx context.Context, client *http.Client, lastID *string, handler func([]byte) error, userAgent string) (retErr error) {
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
	defer func() {
		if err := closeBody(resp); err != nil {
			retErr = errors.Join(retErr, err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP status: %s", resp.Status)
	}

	reader := bufio.NewScanner(resp.Body)
	reader.Buffer(make([]byte, 4096), 1024*1024)

	var data []string
	var eventID string

	dispatch := func() error {
		if len(data) == 0 {
			return nil
		}

		if err := handler([]byte(strings.Join(data, "\n"))); err != nil {
			return err
		}

		if eventID != "" {
			*lastID = eventID
		}

		data = nil
		return nil
	}

	for reader.Scan() {
		line := reader.Text()

		if line == "" {
			if err := dispatch(); err != nil {
				return err
			}
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

func Listen(ctx context.Context, client *http.Client, lastID *string, handler func([]byte) error, userAgent string) error {
	return consume(ctx, client, lastID, handler, userAgent)
}

func Run(
	ctx context.Context,
	client *http.Client,
	handler func([]byte) error,
	userAgent string,
) error {
	var lastID string
	delay := time.Second

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := Listen(ctx, client, &lastID, handler, userAgent)

		if ctx.Err() != nil {
			return ctx.Err()
		}

		switch {
		case err == nil, errors.Is(err, io.EOF): // Reconectar.
			continue

		case errors.Is(err, io.EOF):
			// Conexión cerrada: reconectar.

		case errors.Is(err, context.Canceled):
			return err

		case errors.Is(err, context.DeadlineExceeded):
			return err

		default:
			var urlErr *url.Error

			if errors.As(err, &urlErr) {
				// Error de conexión HTTP: reconectar.
				break
			}

			// Error desconocido o no recuperable.
			return err
		}

		timer := time.NewTimer(delay)

		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}

		delay *= 2
		if delay > 30*time.Second {
			delay = 30 * time.Second
		}
	}
}
