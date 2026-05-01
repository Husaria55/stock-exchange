package repository

import (
	"context"
	"database/sql"

	"github.com/Husaria55/stock-exchange/internal/domain"
)

type PostgresRepo struct {
	db *sql.DB
}

func NewPostgresRepo(db *sql.DB) *PostgresRepo {
	return &PostgresRepo{db: db}
}

func (r *PostgresRepo) BuyStock(ctx context.Context, walletID, stockName string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `INSERT INTO wallets (id) VALUES ($1) ON CONFLICT DO NOTHING`, walletID)
	if err != nil {
		return err
	}

	var bankQty int
	err = tx.QueryRowContext(ctx, `SELECT quantity FROM bank_stocks WHERE name = $1 FOR UPDATE`, stockName).Scan(&bankQty)
	if err == sql.ErrNoRows {
		return domain.ErrStockNotFound
	} else if err != nil {
		return err
	}

	if bankQty <= 0 {
		return domain.ErrInsufficientBank
	}

	_, err = tx.ExecContext(ctx, `UPDATE bank_stocks SET quantity = quantity - 1 WHERE name = $1`, stockName)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO wallet_stocks (wallet_id, stock_name, quantity) 
		VALUES ($1, $2, 1) 
		ON CONFLICT (wallet_id, stock_name) 
		DO UPDATE SET quantity = wallet_stocks.quantity + 1`,
		walletID, stockName,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO audit_log (type, wallet_id, stock_name) VALUES ('buy', $1, $2)`, walletID, stockName)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepo) SellStock(ctx context.Context, walletID, stockName string) error {
	tx, err := r.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var walletQty int
	err = tx.QueryRowContext(ctx, `SELECT quantity FROM wallet_stocks WHERE wallet_id = $1 AND stock_name = $2 FOR UPDATE`, walletID, stockName).Scan(&walletQty)
	if err == sql.ErrNoRows || walletQty <= 0 {
		return domain.ErrInsufficientWallet
	} else if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `UPDATE wallet_stocks SET quantity = quantity - 1 WHERE wallet_id = $1 AND stock_name = $2`, walletID, stockName)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO bank_stocks (name, quantity) 
		VALUES ($1, 1) 
		ON CONFLICT (name) 
		DO UPDATE SET quantity = bank_stocks.quantity + 1`,
		stockName,
	)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `INSERT INTO audit_log (type, wallet_id, stock_name) VALUES ('sell', $1, $2)`, walletID, stockName)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PostgresRepo) SetBankState(ctx context.Context, stocks []domain.Stock) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `UPDATE bank_stocks SET quantity = 0`)
	if err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO bank_stocks (name, quantity) 
		VALUES ($1, $2) 
		ON CONFLICT (name) DO UPDATE SET quantity = EXCLUDED.quantity`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range stocks {
		if _, err := stmt.ExecContext(ctx, s.Name, s.Quantity); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepo) GetBankStocks(ctx context.Context) ([]domain.Stock, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT name, quantity FROM bank_stocks WHERE quantity > 0 ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []domain.Stock
	for rows.Next() {
		var s domain.Stock
		if err := rows.Scan(&s.Name, &s.Quantity); err != nil {
			return nil, err
		}
		stocks = append(stocks, s)
	}
	return stocks, rows.Err()
}

func (r *PostgresRepo) GetWallet(ctx context.Context, walletID string) (*domain.Wallet, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT stock_name, quantity FROM wallet_stocks WHERE wallet_id = $1 AND quantity > 0 ORDER BY stock_name ASC`, walletID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	wallet := &domain.Wallet{ID: walletID, Stocks: []domain.Stock{}}
	for rows.Next() {
		var s domain.Stock
		if err := rows.Scan(&s.Name, &s.Quantity); err != nil {
			return nil, err
		}
		wallet.Stocks = append(wallet.Stocks, s)
	}
	return wallet, rows.Err()
}

func (r *PostgresRepo) GetWalletStockQuantity(ctx context.Context, walletID, stockName string) (int, error) {
	var qty int
	err := r.db.QueryRowContext(ctx, `SELECT quantity FROM wallet_stocks WHERE wallet_id = $1 AND stock_name = $2`, walletID, stockName).Scan(&qty)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return qty, err
}

func (r *PostgresRepo) GetAuditLog(ctx context.Context) ([]domain.AuditLogEntry, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT type, wallet_id, stock_name FROM audit_log ORDER BY created_at ASC LIMIT 10000`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.AuditLogEntry
	for rows.Next() {
		var l domain.AuditLogEntry
		if err := rows.Scan(&l.Type, &l.WalletID, &l.StockName); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	
	if logs == nil {
		logs = make([]domain.AuditLogEntry, 0)
	}
	
	return logs, rows.Err()
}