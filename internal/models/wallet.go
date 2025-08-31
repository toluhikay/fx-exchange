package models

import "time"

type Wallet struct {
	ID            string             `json:"id"`
	EmailOrMobile string             `json:"email"`
	UserId        string             `json:"user_id"`
	Balances      map[string]float64 `json:"balances"`
	CreatedAt     time.Time          `json:"created_at"`
}

type DepositRequest struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
}

type SwapRequest struct {
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	Amount       float64 `json:"amount"`
}

type TransferRequest struct {
	ReceiverID string  `json:"receiver_id"`
	Currency   string  `json:"currency"`
	Amount     float64 `json:"amount"`
}

type Transaction struct {
	ID              string    `json:"id"`
	WalletID        string    `json:"wallet_id"`
	Type            string    `json:"type"`
	FromCurrency    *string   `json:"from_currency"`
	ToCurrency      *string   `json:"to_currency"`
	Amount          *float64  `json:"amount"`
	ConvertedAmount *float64  `json:"converted_amount"`
	Rate            *float64  `json:"rate"`
	Timestamp       time.Time `json:"timestamp"`
}

type BalanceResponse struct {
	Balances map[string]float64 `json:"balances"`
	TotalUSD float64            `json:"total_usd"`
}
