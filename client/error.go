package client

import "errors"

var (
	ErrRegAccount     = errors.New("client: failed to register an account")
	ErrUpdateAccount  = errors.New("client: failed to update the account data")
	ErrEncodeAccount  = errors.New("client: failed to encode account data")
	ErrDecodeAccount  = errors.New("client: failed to decode account data")
	ErrGetAccountData = errors.New("client: failed to get the account data")
)
