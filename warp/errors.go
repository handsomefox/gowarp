package warp

import "errors"

var (
	ErrCreateRequest    = errors.New("error creating request")
	ErrRegisterAccount  = errors.New("error when registering account, maybe rate limited by cf")
	ErrDecodeResponse   = errors.New("error decoding response")
	ErrSetKey           = errors.New("error setting account key")
	ErrDeleteAccount    = errors.New("error deleting account")
	ErrAddReferrer      = errors.New("error adding referrer")
	ErrFetchAccData     = errors.New("error fetching account data")
	ErrDecodeAccData    = errors.New("error decoding account data from json")
	ErrStashedDataEmpty = errors.New("this stashed entry is nil")
	ErrStashNotFilled   = errors.New("stash is not yet filled")
)
