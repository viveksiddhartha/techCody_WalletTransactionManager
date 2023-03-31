package handlers

import (
	"context"
	"encoding/json"
	"mfus_WalletTransactionManager/models"
	"mfus_WalletTransactionManager/services"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func CreateAccountHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request body
		var request models.Account
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Invalid request body"})
			return
		}
		//Check if duplicate request is received
		account, _ := services.GetAccountByEmail(client, request.Email)
		if account != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Account already exist"})
			return
		}

		// Create new account document
		newAccount := models.Account{
			Email:          request.Email,
			Type:           request.Type,
			Balance:        request.Balance,
			HoldBalance:    0,
			Transactions:   []models.Transaction{},
			CreatedAt:      time.Now(),
			DateModified:   time.Time{},
			VirtualWallets: []string{},
		}

		// Insert new account document into database
		result, err := client.Database("walletManager").Collection("accounts").InsertOne(context.Background(), newAccount)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Failed to create account"})
			return
		}

		// Return success response with new account ID
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(models.SuccessResponse{
			Message: "Account created successfully",
			Data:    result.InsertedID,
		})
	}
}

func GetAccountHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse account ID from URL parameter
		vars := mux.Vars(r)
		accountID, err := primitive.ObjectIDFromHex(vars["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Invalid account ID"})
			return
		}
		// Find account document in database
		var account models.Account
		err = client.Database("walletManager").Collection("accounts").FindOne(context.Background(), bson.M{"_id": accountID}).Decode(&account)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Account not found"})
			return
		}

		// Return success response with account information
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.SuccessResponse{
			Message: "Account found",
			Data:    account,
		})
	}
}

// Handler for creating a new virtual wallet transaction
func CreateTransactionHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse virtual wallet ID from URL path parameter
		vars := mux.Vars(r)
		virtualWalletID, err := primitive.ObjectIDFromHex(vars["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Invalid virtual wallet ID"})
			return
		}
		// Decode request body
		var reqBody models.CreateTransactionRequest
		err = json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Invalid request body"})
			return
		}

		// Validate transaction type
		if reqBody.Type != "debit" && reqBody.Type != "credit" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Invalid transaction type. Must be 'debit' or 'credit'"})
			return
		}

		// Validate transaction amount
		if reqBody.Amount <= 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Transaction amount must be positive"})
			return
		}

		// Find virtual wallet document in database
		virtualWallet, err := services.FindVirtualWallet(client, virtualWalletID, "transactions")
		if err != nil {
			if err == mongo.ErrNoDocuments {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Virtual wallet not found"})
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Failed to retrieve virtual wallet"})
			}
			return
		}

		// Update virtual wallet balance and add transaction
		if reqBody.Type == "debit" {
			if reqBody.Amount > virtualWallet.Balance {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Insufficient balance"})
				return
			}
			virtualWallet.Balance -= reqBody.Amount
		} else {
			virtualWallet.Balance += reqBody.Amount
		}
		virtualWallet.Transactions = append(virtualWallet.Transactions, models.Transaction{
			Type:      models.TransactionType(reqBody.Type),
			Amount:    reqBody.Amount,
			CreatedAt: time.Now(),
		})
		virtualWallet.DateModified = time.Now()

		_, err = client.Database("walletManager").Collection("virtual_wallets").UpdateOne(
			context.Background(),
			bson.M{"_id": virtualWalletID},
			bson.M{"$set": bson.M{
				"balance":       virtualWallet.Balance,
				"transactions":  virtualWallet.Transactions,
				"date_modified": virtualWallet.DateModified,
			}},
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Failed to update virtual wallet"})
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.SuccessResponse{
			Message: "Transaction created successfully",
		})
	}
}

// Handler function to get the total balance for a customer across all virtual wallets
func GetCustomerTotalBalanceHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse customer ID from query parameter
		customerID := r.URL.Query().Get("customer_id")
		// Get total balance for customer across all virtual wallets
		totalBalance, err := GetCustomerTotalBalance(client, customerID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Failed to retrieve customer balance"})
			return
		}

		// Return success response with customer total balance
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.SuccessResponse{
			Message: "Customer total balance retrieved successfully",
			Data:    totalBalance,
		})
	}
}

// Helper function to get total balance for a customer across all virtual wallets
func GetCustomerTotalBalance(client *mongo.Client, customerID string) (float64, error) {
	virtualWallets, err := FindAllVirtualWallets(client, customerID)
	if err != nil {
		return 0, err
	}
	var totalBalance float64
	for _, virtualWallet := range virtualWallets {
		totalBalance += virtualWallet.Balance
	}

	return totalBalance, nil
}

func FindAllVirtualWallets(client *mongo.Client, customerID string) ([]models.VirtualWallet, error) {
	// Build filter query
	filter := bson.M{"customer_id": customerID}
	// Find all virtual wallets that match filter in database
	cursor, err := client.Database("walletManager").Collection("virtual_wallets").Find(context.Background(), filter)
	if err != nil {
		return []models.VirtualWallet{}, err
	}
	defer cursor.Close(context.Background())

	// Build slice of virtual wallets from cursor
	var virtualWallets []models.VirtualWallet
	for cursor.Next(context.Background()) {
		var virtualWallet models.VirtualWallet
		err := cursor.Decode(&virtualWallet)
		if err != nil {
			return []models.VirtualWallet{}, err
		}
		virtualWallets = append(virtualWallets, virtualWallet)
	}

	return virtualWallets, nil
}
