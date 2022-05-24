package warpgen

import "errors"

var (
	ErrCreateRequest   = errors.New("error creating a request")
	ErrRegisterAccount = errors.New("error when registering an account, maybe rate limited by cf")
	ErrDecodeResponse  = errors.New("error decoding response")
	ErrSetKey          = errors.New("error when setting account key")
	ErrDeleteAccount   = errors.New("error when deleting account")
	ErrAddReferrer     = errors.New("error when adding referrer")
	ErrFetchAccData    = errors.New("error when fetching account data")
	ErrDecodeAccData   = errors.New("error when deconding account data from json")
)
