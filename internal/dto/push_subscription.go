package dto

type PushSubscribeRequest struct {
	Endpoint string `json:"endpoint" binding:"required"`
	P256dh   string `json:"p256dh" binding:"required"`
	Auth     string `json:"auth" binding:"required"`
}

type PushUnsubscribeRequest struct {
	Endpoint string `json:"endpoint" binding:"required"`
}

type PushVAPIDKeyResponse struct {
	PublicKey string `json:"public_key"`
}
