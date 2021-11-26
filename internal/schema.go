package internal

type AddressResponse struct {
	Address string `json:"address"`
}

type MessageCountResponse struct {
	MessageCount int64 `json:"messageCount"`
}

type SecondsLeftResponse struct {
	SecondsLeft int64 `json:"secondsLeft"`
}

type ExpiredResponse struct {
	Expired bool `json:"expired"`
}

type ResetResponse struct {
	Response string `json:"Response"`
}
