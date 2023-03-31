package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// VirtualWallet represents a virtual wallet document in MongoDB
type VirtualWallet struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	CustomerID   string             `bson:"customer_id"`
	WalletType   WalletType         `bson:"WalletType"`
	Balance      float64            `bson:"balance"`
	HoldBalance  float64            `bson:"hold_balance"`
	Transactions []Transaction      `bson:"transactions,omitempty"`
	DateCreated  time.Time          `bson:"date_created,omitempty"`
	DateModified time.Time          `bson:"date_modified,omitempty"`
}

type WalletType string

const (
	CashWallet    WalletType = "CashWallet"
	CreditWallet  WalletType = "CreditWallet"
	RewardWallet  WalletType = "RewardWallet"
	TradeWallet   WalletType = "TradeWallet"
	TransitWallet WalletType = "TransitWallet"
)

// Request body for creating a new virtual wallet
type CreateVirtualWalletRequest struct {
	CustomerID string  `json:"customer_id"`
	Balance    float64 `json:"balance"`
}

// Request body for creating a new virtual wallet transaction
type CreateTransactionRequest struct {
	Type   string  `json:"type"`
	Amount float64 `json:"amount"`
}
