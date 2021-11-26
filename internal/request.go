package internal

type ReplyRequest struct {
	Reply struct {
		MessageID string `json:"messageId"`
		ReplyBody string `json:"replyBody"`
	} `json:"Reply"`
}

type ForwardRequest struct {
	Forward struct {
		MessageID      string `json:"messageId"`
		ForwardAddress string `json:"forwardAddress"`
	} `json:"Forward"`
}
