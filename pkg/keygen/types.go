package keygen

import (
	"encoding/json"
	"net/http"
)

type keygen struct {
	accounts *createdAccounts
	client   *http.Client
	writer   http.ResponseWriter
}

type result struct {
	Type     string      `json:"account_type"`
	RefCount json.Number `json:"referral_count"`
	License  string      `json:"license"`
}

type createdAccounts struct {
	First  *account
	Second *account
}

type account struct {
	Id      string  `json:"id"`
	Account license `json:"account"`
	Token   string  `json:"token"`
}

type license struct {
	License string `json:"license"`
}
