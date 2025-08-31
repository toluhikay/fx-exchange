package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/toluhikay/fx-exchange/internal/models"
)

type RepositoryImpl interface {
	CreateWallet(ctx context.Context, emailOrMobile string, user_id uuid.UUID) (*models.Wallet, error)
	GetWallet(ctx context.Context, id string) (*models.Wallet, error)
	Deposit(ctx context.Context, walletID, currency string, amount float64) error
	Swap(ctx context.Context, walletID, fromCurrency, toCurrency string, amount, rate, convertedAmount float64) error
	Transfer(ctx context.Context, senderID, receiverID, fromCurrency, toCurrency string, amount, rate, convertedAmount float64) error
	GetTransactionHistory(ctx context.Context, walletID string) ([]models.Transaction, error)
	GetWalletByUserID(ctx context.Context, id uuid.UUID) (*models.Wallet, error)
	GetBalancesWithUSD(ctx context.Context, walletID string) (*models.BalanceResponse, error)
}

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateWallet(ctx context.Context, email string, user_id uuid.UUID) (*models.Wallet, error) {
	fmt.Println("user id", user_id)
	walletID := uuid.New().String()
	balances := map[string]float64{"cNGN": 0, "cXAF": 0, "USDx": 0, "EURx": 0}
	balancesJSON, err := json.Marshal(balances)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal balances: %w", err)
	}
	query := `INSERT INTO wallets (id, email, user_id, balances, created_at) 
             VALUES ($1, $2, $3, $4, $5) RETURNING created_at`
	var createdAt time.Time
	err = r.db.QueryRowContext(ctx, query, walletID, email, user_id, balancesJSON, time.Now()).Scan(&createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create wallet: %w", err)
	}
	return &models.Wallet{
		ID:            walletID,
		EmailOrMobile: email,
		Balances:      balances,
		CreatedAt:     createdAt,
	}, nil
}

func (r *Repository) GetWallet(ctx context.Context, Id string) (*models.Wallet, error) {
	query := `SELECT id, email, user_id, balances, created_at FROM wallets WHERE id = $1`
	var balancesJSON []byte
	var wallet models.Wallet
	err := r.db.QueryRowContext(ctx, query, Id).Scan(&wallet.ID, &wallet.EmailOrMobile, &wallet.UserId, &balancesJSON, &wallet.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	if err := json.Unmarshal(balancesJSON, &wallet.Balances); err != nil {
		return nil, fmt.Errorf("failed to unmarshal balances: %w", err)
	}
	return &wallet, nil
}

func (r *Repository) GetWalletByUserID(ctx context.Context, id uuid.UUID) (*models.Wallet, error) {
	query := `SELECT id, email, balances, created_at FROM wallets WHERE user_id = $1`
	var balancesJSON []byte
	var wallet models.Wallet
	fmt.Println(id)
	err := r.db.QueryRowContext(ctx, query, id).Scan(&wallet.ID, &wallet.EmailOrMobile, &balancesJSON, &wallet.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet: %w", err)
	}
	if err := json.Unmarshal(balancesJSON, &wallet.Balances); err != nil {
		return nil, fmt.Errorf("failed to unmarshal balances: %w", err)
	}
	return &wallet, nil
}

func (r *Repository) Deposit(ctx context.Context, walletID, currency string, amount float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	query := `SELECT balances FROM wallets WHERE id = $1 FOR UPDATE`
	var balancesJSON []byte
	err = tx.QueryRowContext(ctx, query, walletID).Scan(&balancesJSON)
	if err != nil {
		return fmt.Errorf("failed to lock wallet: %w", err)
	}

	var balances map[string]float64
	if err := json.Unmarshal(balancesJSON, &balances); err != nil {
		return fmt.Errorf("failed to unmarshal balances: %w", err)
	}

	if _, ok := balances[currency]; !ok {
		return fmt.Errorf("unsupported currency: %s", currency)
	}
	if amount <= 0 {
		return fmt.Errorf("invalid deposit amount")
	}

	balances[currency] += amount
	balancesJSON, err = json.Marshal(balances)
	if err != nil {
		return fmt.Errorf("failed to marshal balances: %w", err)
	}

	query = `UPDATE wallets SET balances = $1 WHERE id = $2`
	_, err = tx.ExecContext(ctx, query, balancesJSON, walletID)
	if err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	query = `INSERT INTO transactions (id, wallet_id, type, to_currency, amount, timestamp) 
             VALUES ($1, $2, $3, $4, $5, $6)`
	_, err = tx.ExecContext(ctx, query, uuid.New().String(), walletID, "deposit", currency, amount, time.Now())
	if err != nil {
		return fmt.Errorf("failed to log transaction: %w", err)
	}

	return tx.Commit()
}

func (r *Repository) Swap(ctx context.Context, walletID, fromCurrency, toCurrency string, amount, rate, convertedAmount float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	query := `SELECT balances FROM wallets WHERE id = $1 FOR UPDATE`
	var balancesJSON []byte
	err = tx.QueryRowContext(ctx, query, walletID).Scan(&balancesJSON)
	if err != nil {
		return fmt.Errorf("failed to lock wallet: %w", err)
	}

	var balances map[string]float64
	if err := json.Unmarshal(balancesJSON, &balances); err != nil {
		return fmt.Errorf("failed to unmarshal balances: %w", err)
	}

	if amount <= 0 || balances[fromCurrency] < amount {
		return fmt.Errorf("invalid amount or insufficient balance")
	}
	if _, ok := balances[toCurrency]; !ok {
		return fmt.Errorf("unsupported currency: %s", toCurrency)
	}

	balances[fromCurrency] -= amount
	balances[toCurrency] += convertedAmount
	balancesJSON, err = json.Marshal(balances)
	if err != nil {
		return fmt.Errorf("failed to marshal balances: %w", err)
	}

	query = `UPDATE wallets SET balances = $1 WHERE id = $2`
	_, err = tx.ExecContext(ctx, query, balancesJSON, walletID)
	if err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	query = `INSERT INTO transactions (id, wallet_id, type, from_currency, to_currency, amount, converted_amount, rate, timestamp) 
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err = tx.ExecContext(ctx, query, uuid.New().String(), walletID, "swap", fromCurrency, toCurrency, amount, convertedAmount, rate, time.Now())
	if err != nil {
		return fmt.Errorf("failed to log transaction: %w", err)
	}

	return tx.Commit()
}

func (r *Repository) Transfer(ctx context.Context, senderID, receiverID, fromCurrency, toCurrency string, amount, rate, convertedAmount float64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	query := `SELECT id, balances FROM wallets WHERE id IN ($1, $2) FOR UPDATE`
	rows, err := tx.QueryContext(ctx, query, senderID, receiverID)
	if err != nil {
		return fmt.Errorf("failed to lock wallets: %w", err)
	}
	defer rows.Close()

	var senderWallet, receiverWallet models.Wallet
	foundSender, foundReceiver := false, false
	for rows.Next() {
		var id string
		var balancesJSON []byte
		if err := rows.Scan(&id, &balancesJSON); err != nil {
			return fmt.Errorf("failed to scan wallet: %w", err)
		}
		var balances map[string]float64
		if err := json.Unmarshal(balancesJSON, &balances); err != nil {
			return fmt.Errorf("failed to unmarshal balances: %w", err)
		}
		switch id {
		case senderID:
			senderWallet = models.Wallet{Balances: balances}
			foundSender = true
		case receiverID:
			receiverWallet = models.Wallet{Balances: balances}
			foundReceiver = true
		}
	}
	if !foundSender {
		return fmt.Errorf("sender wallet %s not found", senderID)
	}
	if !foundReceiver {
		return fmt.Errorf("receiver wallet %s not found", receiverID)
	}

	if amount <= 0 || senderWallet.Balances[fromCurrency] < amount {
		return fmt.Errorf("invalid amount or insufficient balance")
	}
	if _, ok := senderWallet.Balances[fromCurrency]; !ok {
		return fmt.Errorf("unsupported sender currency: %s", fromCurrency)
	}
	if _, ok := receiverWallet.Balances[toCurrency]; !ok {
		return fmt.Errorf("unsupported receiver currency: %s", toCurrency)
	}

	senderWallet.Balances[fromCurrency] -= amount
	receiverWallet.Balances[toCurrency] += convertedAmount

	senderBalancesJSON, err := json.Marshal(senderWallet.Balances)
	if err != nil {
		return fmt.Errorf("failed to marshal sender balances: %w", err)
	}
	receiverBalancesJSON, err := json.Marshal(receiverWallet.Balances)
	if err != nil {
		return fmt.Errorf("failed to marshal receiver balances: %w", err)
	}

	query = `UPDATE wallets SET balances = $1 WHERE id = $2`
	_, err = tx.ExecContext(ctx, query, senderBalancesJSON, senderID)
	if err != nil {
		return fmt.Errorf("failed to update sender wallet: %w", err)
	}
	_, err = tx.ExecContext(ctx, query, receiverBalancesJSON, receiverID)
	if err != nil {
		return fmt.Errorf("failed to update receiver wallet: %w", err)
	}

	query = `INSERT INTO transactions (id, wallet_id, type, from_currency, to_currency, amount, converted_amount, rate, timestamp) 
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, err = tx.ExecContext(ctx, query, uuid.New().String(), senderID, "transfer", fromCurrency, toCurrency, amount, convertedAmount, rate, time.Now())
	if err != nil {
		return fmt.Errorf("failed to log transaction: %w", err)
	}

	return tx.Commit()
}

func (r *Repository) GetTransactionHistory(ctx context.Context, walletID string) ([]models.Transaction, error) {
	query := `SELECT id, wallet_id, type, from_currency, to_currency, amount, converted_amount, rate, timestamp 
             FROM transactions WHERE wallet_id = $1 ORDER BY timestamp DESC`
	rows, err := r.db.QueryContext(ctx, query, walletID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction history: %w", err)
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var t models.Transaction
		err := rows.Scan(&t.ID, &t.WalletID, &t.Type, &t.FromCurrency, &t.ToCurrency, &t.Amount, &t.ConvertedAmount, &t.Rate, &t.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, t)
	}
	return transactions, nil
}

func (r *Repository) GetBalancesWithUSD(ctx context.Context, walletID string) (*models.BalanceResponse, error) {
	// Fetch wallet balances
	query := `SELECT balances FROM wallets WHERE id = $1`
	var balancesJSON []byte
	err := r.db.QueryRowContext(ctx, query, walletID).Scan(&balancesJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to get wallet balances: %w", err)
	}
	balances := make(map[string]float64)
	if err := json.Unmarshal(balancesJSON, &balances); err != nil {
		return nil, fmt.Errorf("failed to unmarshal balances: %w", err)
	}

	// Calculate total USD equivalent
	totalUSD := 0.0
	for currency, amount := range balances {
		if currency == "USDx" {
			totalUSD += amount // USDx is already in USD
			continue
		}
		// Get latest rate to USDx
		rateQuery := `SELECT rate FROM fx_rates WHERE from_currency = $1 AND to_currency = 'USDx' ORDER BY timestamp DESC LIMIT 1`
		var rate float64
		err := r.db.QueryRowContext(ctx, rateQuery, currency).Scan(&rate)
		if err != nil {
			// Skip if no rate available (e.g., new currency)
			continue
		}
		totalUSD += amount * rate
	}

	return &models.BalanceResponse{
		Balances: balances,
		TotalUSD: totalUSD,
	}, nil
}
