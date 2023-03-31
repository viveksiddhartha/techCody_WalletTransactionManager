package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Account struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	Email          string             `bson:"email"`
	Type           AccountType        `bson:"type,omitempty"`
	Balance        float64            `bson:"balance,omitempty"`
	HoldBalance    float64            `bson:"hold_balance,omitempty"`
	Transactions   []Transaction      `bson:"transactions,omitempty"`
	CreatedAt      time.Time          `bson:"created_at,omitempty"`
	DateModified   time.Time          `bson:"date_modified"`
	VirtualWallets []string           `bson:"virtual_wallets,omitempty"`
}

type AccountType string

const (
	Retail         AccountType = "Retail"
	Corporate      AccountType = "Corporate"
	ChannelPartner AccountType = "ChannelPartner"
	Traders        AccountType = "Traders"
	PrimeCorporate AccountType = "PrimeCorporate"
)

type Transaction struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Type      TransactionType    `bson:"type,omitempty"`
	Amount    float64            `bson:"amount,omitempty"`
	CreatedAt time.Time          `bson:"created_at,omitempty"`
}

type TransactionType string

const (
	Deposit  TransactionType = "deposit"
	Withdraw TransactionType = "withdraw"
)
