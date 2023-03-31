package models

type CreateAccountRequest struct {
	Type    AccountType `json:"type"`
	Balance float64     `json:"balance"`
}

// Request body for holding funds in a virtual wallet
type HoldRequest struct {
	Amount float64 `json:"amount"`
}

// Request body for releasing funds from a virtual wallet hold balance
type ReleaseHoldBalanceRequest struct {
	Amount float64 `json:"amount"`
}
