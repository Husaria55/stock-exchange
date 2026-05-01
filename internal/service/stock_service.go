package service

import (
	"context"

	"github.com/Husaria55/stock-exchange/internal/domain"
	"github.com/Husaria55/stock-exchange/internal/repository"
)

type StockService interface {
	BuyStock(ctx context.Context, walletID, stockName string) error
	SellStock(ctx context.Context, walletID, stockName string) error
	GetWallet(ctx context.Context, walletID string) (*domain.Wallet, error)
	GetWalletStockQuantity(ctx context.Context, walletID, stockName string) (int, error)
	GetBankStocks(ctx context.Context) ([]domain.Stock, error)
	SetBankState(ctx context.Context, stocks []domain.Stock) error
	GetAuditLog(ctx context.Context) ([]domain.AuditLogEntry, error)
}

type stockService struct {
	repo *repository.PostgresRepo
}

func NewStockService(repo *repository.PostgresRepo) StockService {
	return &stockService{repo: repo}
}

func (s *stockService) BuyStock(ctx context.Context, walletID, stockName string) error {
	return s.repo.BuyStock(ctx, walletID, stockName)
}

func (s *stockService) SellStock(ctx context.Context, walletID, stockName string) error {
	return s.repo.SellStock(ctx, walletID, stockName)
}

func (s *stockService) GetWallet(ctx context.Context, walletID string) (*domain.Wallet, error) {
	return s.repo.GetWallet(ctx, walletID)
}

func (s *stockService) GetWalletStockQuantity(ctx context.Context, walletID, stockName string) (int, error) {
	return s.repo.GetWalletStockQuantity(ctx, walletID, stockName)
}

func (s *stockService) GetBankStocks(ctx context.Context) ([]domain.Stock, error) {
	return s.repo.GetBankStocks(ctx)
}

func (s *stockService) SetBankState(ctx context.Context, stocks []domain.Stock) error {
	return s.repo.SetBankState(ctx, stocks)
}

func (s *stockService) GetAuditLog(ctx context.Context) ([]domain.AuditLogEntry, error) {
	return s.repo.GetAuditLog(ctx)
}