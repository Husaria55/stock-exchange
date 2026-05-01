package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"testing"

	_ "github.com/lib/pq"
	"github.com/Husaria55/stock-exchange/internal/domain"
)

func setupTestDB(t *testing.T) *sql.DB {
	dsn := "host=localhost port=5432 user=stockuser password=stockpassword dbname=stockexchange sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to test db: %v", err)
	}

	_, err = db.Exec(`
		TRUNCATE TABLE audit_log, wallet_stocks, wallets, bank_stocks CASCADE;
	`)
	if err != nil {
		t.Fatalf("Failed to truncate tables: %v", err)
	}

	return db
}

func TestConcurrentBuyStock(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") == "" {
		t.Skip("Skipping integration test; set RUN_INTEGRATION_TESTS=1 to run")
	}

	db := setupTestDB(t)
	defer db.Close()
	repo := NewPostgresRepo(db)
	ctx := context.Background()

	stockName := "TSLA"
	err := repo.SetBankState(ctx, []domain.Stock{{Name: stockName, Quantity: 50}})
	if err != nil {
		t.Fatalf("Failed to set initial bank state: %v", err)
	}

	concurrentRequests := 100
	var wg sync.WaitGroup
	wg.Add(concurrentRequests)

	successCount := 0
	failCount := 0
	var mu sync.Mutex 

	for i := 0; i < concurrentRequests; i++ {
		go func(workerID int) {
			defer wg.Done()
			walletID := fmt.Sprintf("wallet_%d", workerID)
			
			err := repo.BuyStock(context.Background(), walletID, stockName)
			
			mu.Lock()
			if err == nil {
				successCount++
			} else {
				failCount++
			}
			mu.Unlock()
		}(i)
	}

	wg.Wait()


	if successCount != 50 {
		t.Errorf("Expected exactly 50 successful buys, got %d", successCount)
	}
	if failCount != 50 {
		t.Errorf("Expected exactly 50 failed buys, got %d", failCount)
	}

	bankStocks, _ := repo.GetBankStocks(ctx)
	if len(bankStocks) > 0 {
		t.Errorf("Expected bank to be empty, but found stocks")
	}

	logs, _ := repo.GetAuditLog(ctx)
	if len(logs) != 50 {
		t.Errorf("Expected 50 audit log entries, got %d", len(logs))
	}
}