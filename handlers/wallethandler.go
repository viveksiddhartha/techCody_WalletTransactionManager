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

func HoldBalanceHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse account ID from URL parameter
		vars := mux.Vars(r)
		accountID, err := primitive.ObjectIDFromHex(vars["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Invalid account ID"})
			return
		}
		// Parse request body
		var request models.HoldRequest
		err = json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Invalid request body"})
			return
		}

		// Update account document with hold balance
		update := bson.M{
			"$inc": bson.M{"hold_balance": request.Amount},
		}
		_, err = client.Database("walletManager").Collection("accounts").UpdateOne(context.Background(), bson.M{"_id": accountID}, update)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Failed to hold balance"})
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.SuccessResponse{
			Message: "Balance held successfully",
			Data:    nil,
		})
	}
}

// Handler function to release funds from a virtual wallet hold balance
func ReleaseHoldBalanceHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse virtual wallet ID from URL parameter
		vars := mux.Vars(r)
		virtualWalletID, err := primitive.ObjectIDFromHex(vars["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Invalid virtual wallet ID"})
			return
		}
		// Parse customer ID from query parameter
		customerID := r.URL.Query().Get("customer_id")

		// Parse request body
		var request models.ReleaseHoldBalanceRequest
		err = json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Invalid request body"})
			return
		}

		// Create new virtual wallet transaction to release funds from hold balance
		err = services.CreateVirtualWalletTransaction(client, virtualWalletID, customerID, "release", request.Amount)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: err.Error()})
			return
		}

		// Return success response with transaction information
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(models.SuccessResponse{
			Message: "Hold balance released successfully",
			Data:    request,
		})
	}
}

func GetVirtualWalletTransactionsHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse virtual wallet ID from URL parameter
		vars := mux.Vars(r)
		virtualWalletID, err := primitive.ObjectIDFromHex(vars["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Invalid virtual wallet ID"})
			return
		}

		// Parse customer ID from query parameter
		customerID := r.URL.Query().Get("customer_id")

		// Build filter query
		filter := bson.M{"_id": virtualWalletID}
		if customerID != "" {
			filter["customer_id"] = customerID
		}

		// Find virtual wallet document in database
		var virtualWallet models.VirtualWallet
		err = client.Database("walletManager").Collection("virtual_wallets").FindOne(context.Background(), filter).Decode(&virtualWallet)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Virtual wallet not found"})
			return
		}

		// Return success response with virtual wallet transactions
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.SuccessResponse{
			Message: "Virtual wallet transactions retrieved successfully",
			Data:    virtualWallet.Transactions,
		})
	}
}

// Handler for creating a new virtual wallet
func CreateVirtualWalletHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode request body
		var reqBody models.CreateVirtualWalletRequest
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Invalid request body"})
			return
		}
		// Validate customer ID
		if reqBody.CustomerID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Customer ID is required"})
			return
		}

		// Validate balance
		if reqBody.Balance < 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Balance cannot be negative"})
			return
		}

		// Create new virtual wallet document
		virtualWallet := models.VirtualWallet{
			CustomerID:   reqBody.CustomerID,
			Balance:      reqBody.Balance,
			DateCreated:  time.Now(),
			DateModified: time.Now(),
		}

		// Insert virtual wallet document into database
		result, err := client.Database("walletManager").Collection("virtual_wallets").InsertOne(context.Background(), virtualWallet)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Failed to create virtual wallet"})
			return
		}

		// Return success response with new virtual wallet ID
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(models.SuccessResponse{
			Message: "Virtual wallet created successfully",
			Data:    result.InsertedID,
		})
	}
}

// Handler for retrieving all virtual wallets
func GetAllVirtualWalletsHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Retrieve all virtual wallet documents from database
		cursor, err := client.Database("walletManager").Collection("virtual_wallets").Find(context.Background(), bson.M{})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Failed to retrieve virtual wallets"})
			return
		}
		// Decode virtual wallet documents into slice
		var virtualWallets []models.VirtualWallet
		err = cursor.All(context.Background(), &virtualWallets)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Failed to decode virtual wallets"})
			return
		}

		// Return success response with virtual wallet documents
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.SuccessResponse{
			Message: "Virtual wallets retrieved successfully",
			Data:    virtualWallets,
		})
	}
}

// Handler for retrieving a virtual wallet by ID
func GetVirtualWalletHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse virtual wallet ID from URL path parameter
		vars := mux.Vars(r)
		virtualWalletID, err := primitive.ObjectIDFromHex(vars["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Invalid virtual wallet ID"})
			return
		}
		// Retrieve virtual wallet document from database
		virtualWallet, err := services.FindVirtualWallet(client, virtualWalletID, "")
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

		// Return success response with virtual wallet document
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.SuccessResponse{
			Message: "Virtual wallet retrieved successfully",
			Data:    virtualWallet,
		})
	}
}

// Handler for updating a virtual wallet by ID
func UpdateVirtualWalletHandler(client *mongo.Client) http.HandlerFunc {
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
		var reqBody models.CreateVirtualWalletRequest
		err = json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Invalid request body"})
			return
		}

		// Validate customer ID
		if reqBody.CustomerID == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Customer ID is required"})
			return
		}

		// Validate balance
		if reqBody.Balance < 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Balance cannot be negative"})
			return
		}

		// Find virtual wallet document in database
		virtualWallet, err := services.FindVirtualWallet(client, virtualWalletID, "")
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

		// Update virtual wallet document
		virtualWallet.CustomerID = reqBody.CustomerID
		virtualWallet.Balance = reqBody.Balance
		virtualWallet.DateModified = time.Now()

		_, err = client.Database("walletManager").Collection("virtual_wallets").UpdateOne(
			context.Background(),
			bson.M{"_id": virtualWalletID},
			bson.M{"$set": virtualWallet},
		)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Failed to update virtual wallet"})
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.SuccessResponse{
			Message: "Virtual wallet updated successfully",
		})
	}
}

// Handler for deleting a virtual wallet by ID
func DeleteVirtualWalletHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse virtual wallet ID from URL path parameter
		vars := mux.Vars(r)
		virtualWalletID, err := primitive.ObjectIDFromHex(vars["id"])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Invalid virtual wallet ID"})
			return
		}
		// Delete virtual wallet document from database
		result, err := client.Database("walletManager").Collection("virtual_wallets").DeleteOne(context.Background(), bson.M{"_id": virtualWalletID})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Failed to delete virtual wallet"})
			return
		}
		if result.DeletedCount == 0 {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(models.ErrorResponse{Message: "Virtual wallet not found"})
			return
		}

		// Return success response
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(models.SuccessResponse{
			Message: "Virtual wallet deleted successfully",
		})
	}
}

// To DO Need to test
func GetVirtualWalletTransactionsByTrnTypeDtRangeHandler(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse request parameters
		vars := mux.Vars(r)
		virtualWalletID, err := primitive.ObjectIDFromHex(vars["id"])
		if err != nil {
			http.Error(w, "Invalid virtual wallet ID", http.StatusBadRequest)
			return
		}
		filter := r.URL.Query().Get("type")
		startDateStr := r.URL.Query().Get("start_date")
		endDateStr := r.URL.Query().Get("end_date")

		// Find virtual wallet document by ID
		virtualWallet, err := services.FindVirtualWallet(client, virtualWalletID, "")
		if err != nil {
			http.Error(w, "Virtual wallet not found", http.StatusNotFound)
			return
		}

		// Filter transactions by type
		var filteredTransactions []models.Transaction
		if filter != "" {
			for _, transaction := range virtualWallet.Transactions {
				filteredTransactions = append(filteredTransactions, transaction)
			}
		} else {
			filteredTransactions = virtualWallet.Transactions
		}

		// Filter transactions by date range
		var startDate, endDate time.Time
		if startDateStr != "" {
			startDate, err = time.Parse("2006-01-02", startDateStr)
			if err != nil {
				http.Error(w, "Invalid start date", http.StatusBadRequest)
				return
			}
		} else {
			startDate = time.Time{}
		}
		if endDateStr != "" {
			endDate, err = time.Parse("2006-01-02", endDateStr)
			if err != nil {
				http.Error(w, "Invalid end date", http.StatusBadRequest)
				return
			}
		} else {
			endDate = time.Now()
		}

		// Filter transactions by date range
		var dateFilteredTransactions []models.Transaction
		for _, transaction := range filteredTransactions {
			if models.TransactionType(transaction.Type) == transactionType(filter) &&
				startDate.Before(transaction.CreatedAt) && endDate.After(transaction.CreatedAt) {
				dateFilteredTransactions = append(dateFilteredTransactions, transaction)
			}
		}

		// Marshal transactions to JSON and write response
		jsonData, err := json.Marshal(dateFilteredTransactions)
		if err != nil {
			http.Error(w, "Failed to encode JSON", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)
	}
}

func transactionType(str string) models.TransactionType {
	switch str {
	case "DebitTransaction":
		return models.Deposit
	case "CreditTransaction":
		return models.Withdraw
	default:
		return models.Deposit
	}
}
