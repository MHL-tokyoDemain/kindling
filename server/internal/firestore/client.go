package firestore

import "context"

type Client struct{}

func NewClient(ctx context.Context, credsPath string) (*Client, error) {
	return &Client{}, nil
}

func (c *Client) Close() error {
	return nil
}
