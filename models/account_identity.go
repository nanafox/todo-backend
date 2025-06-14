package models

type AccountIdentity struct {
	ID       string `json:"id"`
	Provider string `json:"provider"`
	UserID   uint
	User     User `json:"user"`
}
