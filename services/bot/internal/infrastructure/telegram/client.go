package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const apiBaseURL = "https://api.telegram.org"

type Client struct {
	token      string
	httpClient *http.Client
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		httpClient: &http.Client{
			Timeout: 35 * time.Second,
		},
	}
}

type Update struct {
	ID      int64    `json:"update_id"`
	Message *Message `json:"message"`
}

type Message struct {
	Text    string   `json:"text"`
	Chat    Chat     `json:"chat"`
	From    *User    `json:"from"`
	Contact *Contact `json:"contact"`
}

type Chat struct {
	ID int64 `json:"id"`
}

type User struct {
	Username string `json:"username"`
}

type Contact struct {
	PhoneNumber string `json:"phone_number"`
}

func (c *Client) GetUpdates(ctx context.Context, offset int64, timeout int) ([]Update, error) {
	values := url.Values{}
	values.Set("timeout", strconv.Itoa(timeout))
	if offset > 0 {
		values.Set("offset", strconv.FormatInt(offset, 10))
	}

	var response struct {
		OK          bool     `json:"ok"`
		Result      []Update `json:"result"`
		Description string   `json:"description"`
	}

	if err := c.get(ctx, "getUpdates", values, &response); err != nil {
		return nil, err
	}
	if !response.OK {
		return nil, fmt.Errorf("telegram getUpdates failed: %s", response.Description)
	}

	return response.Result, nil
}

func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	payload := map[string]any{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "HTML",
	}

	var response struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}

	if err := c.post(ctx, "sendMessage", payload, &response); err != nil {
		return err
	}
	if !response.OK {
		return fmt.Errorf("telegram sendMessage failed: %s", response.Description)
	}

	return nil
}

func (c *Client) get(ctx context.Context, method string, values url.Values, target any) error {
	endpoint := fmt.Sprintf("%s/bot%s/%s?%s", apiBaseURL, c.token, method, values.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, http.NoBody)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return json.NewDecoder(resp.Body).Decode(target)
}

func (c *Client) post(ctx context.Context, method string, payload, target any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("%s/bot%s/%s", apiBaseURL, c.token, method)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return json.NewDecoder(resp.Body).Decode(target)
}
