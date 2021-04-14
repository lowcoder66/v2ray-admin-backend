package response

import "time"

type AccessToken struct {
	Token    string    `json:"token"`
	ReqTime  time.Time `json:"reqTime"`
	ExpireIn int       `json:"expireIn"`
}
