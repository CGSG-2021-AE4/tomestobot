package bxclient

import (
	"fmt"
	"io"

	"tomestobot/pkg/gobx/bxtypes"

	"resty.dev/v3"
)

type BxClient interface {
	SetDebug(b bool)
	Do(method string, bodyData any, respData any) (*resty.Response, error)
	io.Closer
}

type bxClient struct {
	client *resty.Client
	apiUrl string // URL to the API (includes user id and hook)
}

func New(hostUrl string, userId int, secret string) BxClient {
	return &bxClient{
		client: resty.New(),
		apiUrl: fmt.Sprintf("https://%s/rest/%d/%s/", hostUrl, userId, secret),
	}
}

func (c *bxClient) SetDebug(b bool) {
	c.client.SetDebug(b)
}

// From Do functions all errors already wrapped!!! so I do not have to wrap them later to mark error's level
func (c *bxClient) Do(method string, bodyData any, respData any) (*resty.Response, error) {
	// Setup request

	req := c.client.R().
		SetContentType("application/json").
		SetHeader("Accept", "application/json").
		SetBody(bodyData).
		SetError(&bxtypes.ResponseError{})
	if respData != nil {
		req.SetResult(respData)
	}

	// Make request
	resp, err := req.Post(c.apiUrl + method)

	// Handling errors
	if err != nil { // Resty internal error
		return nil, bxtypes.ErrorResty{Err: err}
	}
	if resp.IsError() { // HTTP status code >= 400
		return resp, bxtypes.ErrorStatusCode(resp.StatusCode())
	}

	return resp, nil
}

func (c *bxClient) Close() error {
	return c.client.Close()
}
