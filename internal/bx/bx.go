package bx

import (
	"fmt"

	"tomestobot/api"

	"tomestobot/pkg/gobx/bxclient"
	"tomestobot/pkg/gobx/bxtypes"
)

type bxWrapper struct {
	client bxclient.BxClient
}

func New(crmUrl string, botUserId int, hook string) (api.BxWrapper, error) {
	c := bxclient.New(crmUrl, botUserId, hook)

	// For debug
	// c.SetInsecureSSL(true)
	// c.SetDebug(true)

	return &bxWrapper{
		client: c,
	}, nil
}

func (b *bxWrapper) AuthUserByPhone(phone string) (api.BxUser, error) {
	// Make request
	resp, err := b.client.Do(
		"user.get",
		bxtypes.ReqUserGet{
			Filter: map[string]string{
				"PERSONAL_MOBILE": phone,
			},
		},
		&bxtypes.ArrayResponse[bxtypes.User]{})

	// Check for result to be valid
	if err != nil {
		return nil, fmt.Errorf("during request: %w", err)
	}
	res, ok := resp.(*bxtypes.ArrayResponse[bxtypes.User])
	if !ok {
		return nil, fmt.Errorf("failed to parse response")
	}
	if res.Total == 0 {
		return nil, fmt.Errorf("no such users found")
	}
	if res.Total > 1 {
		return nil, fmt.Errorf("found several users")
	}

	// Create new user
	return &bxUser{
		bx: b.client,
		id: res.Result[0].Id,
	}, nil
}

func (b *bxWrapper) Close() error {
	return b.client.Close()
}
