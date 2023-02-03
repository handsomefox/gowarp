package client

// Account represents a registered CF account.
type Account struct {
	ID      string `json:"id"`
	Account struct {
		License string `json:"license"`
	} `json:"account"`
	Token string `json:"token"`
}

