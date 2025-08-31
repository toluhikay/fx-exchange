package services

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/toluhikay/fx-exchange/internal/fx"
	"github.com/toluhikay/fx-exchange/internal/models"
	"github.com/toluhikay/fx-exchange/internal/repository"
)

type Service struct {
	repo repository.RepositoryImpl
	fx   fx.FXProvider
	mu   sync.Mutex
}

func NewService(repo *repository.Repository, fx fx.FXProvider) *Service {
	return &Service{repo: repo, fx: fx}
}

func (s *Service) CreateWallet(ctx context.Context, email string, userId uuid.UUID) (*models.Wallet, error) {
	return s.repo.CreateWallet(ctx, email, userId)
}

func (s *Service) GetWalletByUserId(ctx context.Context, userId uuid.UUID) (*models.Wallet, error) {
	return s.repo.GetWalletByUserID(ctx, userId)
}

func (s *Service) Deposit(ctx context.Context, walletID, currency string, amount float64) error {
	return s.repo.Deposit(ctx, walletID, currency, amount)
}

func (s *Service) Swap(ctx context.Context, walletID, fromCurrency, toCurrency string, amount float64) (float64, float64, error) {
	fmt.Println(s.fx)
	rate, err := s.fx.GetRate(ctx, fromCurrency, toCurrency)
	if err != nil {
		fmt.Println(err)
		return 0, 0, fmt.Errorf("failed to get FX rate: %w", err)
	}
	convertedAmount := amount * rate
	err = s.repo.Swap(ctx, walletID, fromCurrency, toCurrency, amount, rate, convertedAmount)
	if err != nil {
		return 0, 0, err
	}
	return convertedAmount, rate, nil
}

func (s *Service) Transfer(ctx context.Context, senderID, receiverID, currency string, amount float64) (float64, float64, error) {
	_, err := s.repo.GetWallet(ctx, senderID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get sender wallet: %w", err)
	}
	receiver, err := s.repo.GetWallet(ctx, receiverID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to get receiver wallet: %w", err)
	}

	receiverCurrency := currency
	rate := 1.0
	convertedAmount := amount
	if _, ok := receiver.Balances[currency]; !ok {
		for c := range receiver.Balances {
			receiverCurrency = c
			break
		}
		if currency != receiverCurrency {
			rate, err = s.fx.GetRate(ctx, currency, receiverCurrency)
			if err != nil {
				return 0, 0, fmt.Errorf("failed to get FX rate: %w", err)
			}
			convertedAmount = amount * rate
		}
	}

	err = s.repo.Transfer(ctx, senderID, receiverID, currency, receiverCurrency, amount, rate, convertedAmount)
	if err != nil {
		return 0, 0, err
	}
	return convertedAmount, rate, nil
}

func (s *Service) GetTransactionHistory(ctx context.Context, walletID string) ([]models.Transaction, error) {
	return s.repo.GetTransactionHistory(ctx, walletID)
}

func (s *Service) GetBalancesWithUSD(ctx context.Context, walletID string) (*models.BalanceResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.repo.GetBalancesWithUSD(ctx, walletID)
}
