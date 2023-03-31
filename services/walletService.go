package services

import (
	"context"
	"errors"
	"mfus_WalletTransactionManager/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Helper function for finding a virtual wallet document by ID and optionally retrieving its transactions
func FindVirtualWallet(client *mongo.Client, virtualWalletID primitive.ObjectID, fields string) (*models.VirtualWallet, error) {
	// Define projection to include transaction field if specified
	projection := bson.M{}
	if fields != "" {
		projection[fields] = 1
	}
	// Find virtual wallet document in database
	var virtualWallet models.VirtualWallet
	err := client.Database("walletManager").Collection("virtual_wallets").FindOne(
		context.Background(),
		bson.M{"_id": virtualWalletID},
		options.FindOne().SetProjection(projection),
	).Decode(&virtualWallet)
	if err != nil {
		return nil, err
	}

	return &virtualWallet, nil
}

// Helper function to create a new virtual wallet transaction and update virtual wallet balance
func CreateVirtualWalletTransaction(client *mongo.Client, virtualWalletID primitive.ObjectID, customerID string, transactionType models.TransactionType, amount float64) error {
	// Find virtual wallet document in database
	virtualWallet, err := FindVirtualWallet(client, virtualWalletID, customerID)
	if err != nil {
		return err
	}

	/* 	switch transactionType {
	   	case "debit":
	   		transactionType = "Deposit"
	   	case "credit":
	   		transactionType = "withdraw"
	   	default:
	   		errors.New("Invalid transaction type") */

	// Create new transaction document

	newTransaction := models.Transaction{
		Type:      transactionType,
		Amount:    amount,
		CreatedAt: time.Now(),
	}

	// Build update query
	update := bson.M{
		"$push": bson.M{"transactions": newTransaction},
	}

	// Update virtual wallet document based on transaction type
	switch transactionType {
	case "deposit":
		update["$inc"] = bson.M{"balance": amount}

	case "withdraw":
		if virtualWallet.Balance < amount {
			return errors.New("Insufficient funds")
		}

		update["$inc"] = bson.M{"balance": -amount}

	case "hold":
		if virtualWallet.Balance < amount {
			return errors.New("Insufficient funds")
		}

		update["$inc"] = bson.M{"balance": -amount, "hold_balance": amount}

	case "release":
		if virtualWallet.HoldBalance < amount {
			return errors.New("Invalid amount to release")
		}

		update["$inc"] = bson.M{"balance": amount, "hold_balance": -amount}

	default:
		return errors.New("Invalid transaction type")
	}

	// Update virtual wallet document in database
	err = UpdateVirtualWallet(client, virtualWalletID, customerID, update)
	if err != nil {
		return err
	}

	return nil
}

// Helper function to update virtual wallet document in database by ID and customer ID
func UpdateVirtualWallet(client *mongo.Client, virtualWalletID primitive.ObjectID, customerID string, update bson.M) error {
	// Build filter query
	filter := bson.M{"_id": virtualWalletID}
	if customerID != "" {
		filter["customer_id"] = customerID
	}
	// Update virtual wallet document in database
	_, err := client.Database("walletManager").Collection("virtual_wallets").UpdateOne(context.Background(), filter, update)
	if err != nil {
		return err
	}

	return nil
}
