package models

type PaymentRequest struct {
	Amount     int32  `json:"amount" binding:"required"`
	CancelUrl  string `json:"cancelUrl" binding:"required"`
	Currency   string `json:"currency" binding:"required"`
	Name       string `json:"name" binding:"required"`
	Quantity   int32  `json:"quantity" binding:"required"`
	SuccessUrl string `json:"successUrl" binding:"required"`
}
