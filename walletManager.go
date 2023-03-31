package main

import (
	"context"
	"log"
	"mfus_WalletTransactionManager/handlers"
	"net/http"
	"os"

	handle "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Main function to start HTTP server

func main() {
	// Connect to MongoDB
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	// Create a log file
	logfile, err := os.OpenFile("server.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to create log file: %v", err)
	}
	defer logfile.Close()

	defer client.Disconnect(context.Background())
	// Set up router and routes
	r := mux.NewRouter()

	// Create a new validator instance
	//validate := validator.New()

	// Set up account endpoints
	r.HandleFunc("/accounts", handlers.CreateAccountHandler(client)).Methods("POST")
	r.HandleFunc("/accounts/{id}", handlers.GetAccountHandler(client)).Methods("GET")

	// Set up transaction on Account endpoints
	r.HandleFunc("/accounts/{id}/transactions", handlers.CreateTransactionHandler(client)).Methods("POST")
	r.HandleFunc("/accounts/{id}/hold", handlers.HoldBalanceHandler(client)).Methods("POST")
	r.HandleFunc("/accounts/{id}/release", handlers.ReleaseHoldBalanceHandler(client)).Methods("POST")

	// Set up Wallet endpoints
	r.HandleFunc("/virtual_wallets", handlers.CreateVirtualWalletHandler(client)).Methods("POST")
	r.HandleFunc("/virtual_wallets/{id}", handlers.GetVirtualWalletHandler(client)).Methods("GET")
	r.HandleFunc("/virtual_wallets/{id}", handlers.UpdateVirtualWalletHandler(client)).Methods("PUT")
	r.HandleFunc("/virtual_wallets/{id}", handlers.DeleteVirtualWalletHandler(client)).Methods("DELETE")

	// Set up transaction on Wallet endpoints
	r.HandleFunc("/virtual_wallets/{id}/transactions", handlers.CreateTransactionHandler(client)).Methods("POST")
	r.HandleFunc("/virtual_wallets/{id}/transactions", handlers.GetVirtualWalletTransactionsHandler(client)).Methods("GET")

	// Customer total balance endpoints
	r.HandleFunc("/customers/{id}/total_balance", handlers.GetCustomerTotalBalanceHandler(client)).Methods("GET")

	// Wrap the router with logging and validation middleware
	loggedRouter := handle.LoggingHandler(log.Writer(), r)

	//	validatedRouter := handlers.ValidationMiddleware(validate, loggedRouter)

	//	handlers.ValidateRequest(handle.ValidateRequest{Validate: validate})(loggedRouter))

	// Start server
	log.Fatal(http.ListenAndServe(":8080", loggedRouter))
}
