package tgclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// SetWebAppMenuButton makes the button beside the message field open the Mini
// App, so the dashboard is one tap away instead of buried in a menu.
//
// Called over raw HTTP rather than through the bot library: v5.5.1 predates Web
// Apps and has no MenuButton type at all (a grep for WebApp in that module
// returns nothing). One JSON POST is a far smaller price than replacing the
// library the whole bot is built on.
//
// The button is a single global default, so its label cannot follow each user's
// language the way the rest of the UI does. Russian is the operator's own
// language here; a per-user label would need setChatMenuButton on every chat.
func SetWebAppMenuButton(ctx context.Context, token, label, webAppURL string) error {
	if token == "" {
		return fmt.Errorf("set menu button: token is empty")
	}
	if webAppURL == "" {
		return fmt.Errorf("set menu button: web app url is empty")
	}
	if _, err := url.Parse(webAppURL); err != nil {
		return fmt.Errorf("set menu button: bad url %q: %w", webAppURL, err)
	}

	body, err := json.Marshal(map[string]any{
		"menu_button": map[string]any{
			"type":    "web_app",
			"text":    label,
			"web_app": map[string]string{"url": webAppURL},
		},
	})
	if err != nil {
		return err
	}

	// No chat_id: this sets the default for every chat with the bot.
	endpoint := "https://api.telegram.org/bot" + token + "/setChatMenuButton"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("set menu button: %w", err)
	}
	defer resp.Body.Close()

	// Telegram answers 200 with ok:false for a rejected request, so the status
	// code alone does not tell us whether it worked.
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	if err != nil {
		return fmt.Errorf("set menu button: read response: %w", err)
	}
	var parsed struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return fmt.Errorf("set menu button: decode response: %w", err)
	}
	if !parsed.OK {
		// The token never appears here — only Telegram's own words.
		return fmt.Errorf("set menu button: telegram refused: %s", parsed.Description)
	}
	return nil
}
