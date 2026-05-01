package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/Husaria55/stock-exchange/internal/domain"
	"github.com/Husaria55/stock-exchange/internal/service"
)

type StockHandler struct {
	service service.StockService
}

func NewRouter(s service.StockService) *http.ServeMux {
	h := &StockHandler{service: s}
	mux := http.NewServeMux()

	mux.HandleFunc("POST /wallets/{wallet_id}/stocks/{stock_name}", h.handleTrade)
	mux.HandleFunc("GET /wallets/{wallet_id}", h.handleGetWallet)
	mux.HandleFunc("GET /wallets/{wallet_id}/stocks/{stock_name}", h.handleGetWalletStock)
	mux.HandleFunc("GET /stocks", h.handleGetBankStocks)
	mux.HandleFunc("POST /stocks", h.handleSetBankState)
	mux.HandleFunc("GET /log", h.handleGetAuditLog)
	mux.HandleFunc("POST /chaos", h.handleChaos)

	return mux
}

func (h *StockHandler) handleTrade(w http.ResponseWriter, r *http.Request) {
	walletID := r.PathValue("wallet_id")
	stockName := r.PathValue("stock_name")

	var req struct {
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	var err error
	if req.Type == "buy" {
		err = h.service.BuyStock(r.Context(), walletID, stockName)
	} else if req.Type == "sell" {
		err = h.service.SellStock(r.Context(), walletID, stockName)
	} else {
		http.Error(w, "type must be 'buy' or 'sell'", http.StatusBadRequest)
		return
	}

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrStockNotFound):
			http.Error(w, err.Error(), http.StatusNotFound) // 404 if stock doesn't exist
		case errors.Is(err, domain.ErrInsufficientBank), errors.Is(err, domain.ErrInsufficientWallet):
			http.Error(w, err.Error(), http.StatusBadRequest) // 400 for logic failures
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *StockHandler) handleGetWallet(w http.ResponseWriter, r *http.Request) {
	walletID := r.PathValue("wallet_id")
	wallet, err := h.service.GetWallet(r.Context(), walletID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wallet)
}

func (h *StockHandler) handleGetWalletStock(w http.ResponseWriter, r *http.Request) {
	walletID := r.PathValue("wallet_id")
	stockName := r.PathValue("stock_name")

	qty, err := h.service.GetWalletStockQuantity(r.Context(), walletID, stockName)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}


	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "%d", qty)
}

func (h *StockHandler) handleGetBankStocks(w http.ResponseWriter, r *http.Request) {
	stocks, err := h.service.GetBankStocks(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"stocks": stocks})
}

func (h *StockHandler) handleSetBankState(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Stocks []domain.Stock `json:"stocks"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.SetBankState(r.Context(), req.Stocks); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *StockHandler) handleGetAuditLog(w http.ResponseWriter, r *http.Request) {
	logs, err := h.service.GetAuditLog(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"log": logs})
}

func (h *StockHandler) handleChaos(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Chaos endpoint hit. Shutting down forcefully.")
	os.Exit(1)
}