package bxclient

import (
	"fmt"
	"io"

	"tomestobot/pkg/gobx/bxtypes"

	"resty.dev/v3"
)

type BxClient interface {
	SetDebug(b bool)
	DoRaw(method string, bodyData any, respData any) (*resty.Response, error)
	Do(method string, bodyData any, respData any) (any, error)
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

func (c *bxClient) DoRaw(method string, bodyData any, respData any) (*resty.Response, error) {
	req := c.client.R().
		SetContentType("application/json").
		SetHeader("Accept", "application/json").
		SetBody(bodyData).
		SetError(&bxtypes.ResponseError{})

	if respData != nil {
		req.SetResult(respData)
	}

	resp, err := req.Post(c.apiUrl + method)
	if err != nil {
		return nil, fmt.Errorf("error during request: %w", err)
	}

	if resp.IsError() {
		return resp, resp.Error().(*bxtypes.ResponseError)
	}

	return resp, nil
}

func (c *bxClient) Do(method string, bodyData any, respData any) (any, error) {
	resp, err := c.DoRaw(method, bodyData, respData)
	return resp.Result(), err
}

func (c *bxClient) Close() error {
	return c.client.Close()
}
