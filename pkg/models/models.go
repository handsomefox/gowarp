package models

import (
	"encoding/json"
	"errors"
)

type Account struct {
	ID       any         `bson:"_id,omitempty" json:"id,omitempty"`
	Type     string      `bson:"account_type" json:"account_type"`
	RefCount json.Number `bson:"referral_count" json:"referral_count"`
	License  string      `bson:"license" json:"license"`
}

var (
	ErrInvalidKey       = errors.New("models: invalid key provided")
	ErrDeleteFailed     = errors.New("models: couldn't delete entry")
	ErrNoRecord         = errors.New("models: no record found")
	ErrConnectionFailed = errors.New("models: couldn't connect to database")
	ErrPingFailed       = errors.New("models: couldn't ping database")
	ErrInsertFailed     = errors.New("models: couldn't insert an entry to the database")
)
