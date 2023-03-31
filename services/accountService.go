package services

import (
	"context"
	"fmt"
	"mfus_WalletTransactionManager/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetAccountByEmail(client *mongo.Client, email string) (*models.Account, error) {

	// Get a handle to the accounts collection
	collection := client.Database("walletManager").Collection("accounts")

	// Define a filter for the email
	filter := bson.M{"email": email}

	// Find the account that matches the filter
	var account models.Account
	err := collection.FindOne(context.Background(), filter).Decode(&account)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("account not found")
		} else {
			return nil, fmt.Errorf("failed to find account: %s", err)
		}
	}

	return &account, nil
}
