package response

import "time"

type AccessToken struct {
	Token    string    `json:"token"`
	ReqTime  time.Time `json:"reqTime"`
	ExpireIn uint32    `json:"expireIn"`
}
