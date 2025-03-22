package bx

import (
	"fmt"
	"log/slog"

	"github.com/CGSG-2021-AE4/tomestobot/api"

	"github.com/CGSG-2021-AE4/tomestobot/pkg/gobx/bxclient"
	"github.com/CGSG-2021-AE4/tomestobot/pkg/gobx/bxtypes"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

type BxDescriptor struct {
	BxDomain string `validate:"required,fqdn"` // Full Qualified Domain Name
	BxUserId int    `validate:"required"`
	BxHook   string `validate:"required"`
}

type bxWrapper struct {
	logger *slog.Logger

	client bxclient.BxClient
}

func New(logger *slog.Logger, descr BxDescriptor) (api.BxWrapper, error) {
	// Validate descriptor
	if err := validate.Struct(descr); err != nil {
		return nil, fmt.Errorf("bx wrapper descriptor validation: %w", err)
	}

	// Create bxclient
	c := bxclient.New(descr.BxDomain, descr.BxUserId, descr.BxHook)

	// For debug
	c.SetDebug(api.EnableRestyLogs)

	return &bxWrapper{
		logger: logger,
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
	b.logger.Debug("got resp")

	// Check for result to be valid
	if err != nil {
		return nil, err
	}
	res, ok := resp.Result().(*bxtypes.ArrayResponse[bxtypes.User])
	if !ok {
		return nil, api.ErrorParseResponse
	}
	if res.Total == 0 {
		return nil, api.ErrorUserNotFound
	}
	if res.Total > 1 {
		return nil, api.ErrorSeveralUsersFound
	}

	// Create new user
	return &bxUser{
		bx:   b.client,
		user: res.Result[0],
	}, nil
}

func (b *bxWrapper) AuthUserById(id bxtypes.Id) (api.BxUser, error) {
	// Make request
	resp, err := b.client.Do(
		"user.get",
		bxtypes.ReqUserGet{
			Filter: map[string]string{
				"ID": id.String(),
			},
		},
		&bxtypes.ArrayResponse[bxtypes.User]{})

	// Check for result to be valid
	if err != nil {
		return nil, err
	}
	res, ok := resp.Result().(*bxtypes.ArrayResponse[bxtypes.User])
	if !ok {
		return nil, api.ErrorParseResponse
	}
	if res.Total == 0 {
		return nil, api.ErrorUserNotFound
	}
	if res.Total > 1 {
		return nil, api.ErrorSeveralUsersFound
	}

	// Create new user
	return &bxUser{
		bx:   b.client,
		user: res.Result[0],
	}, nil
}

func (b *bxWrapper) Close() error {
	return b.client.Close()
}
