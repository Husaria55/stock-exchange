package domain

import "errors"

var (
	ErrStockNotFound      = errors.New("stock not found in bank")
	ErrInsufficientBank   = errors.New("insufficient stock in bank")
	ErrInsufficientWallet = errors.New("insufficient stock in wallet")
)