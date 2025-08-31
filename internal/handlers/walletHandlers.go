package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/toluhikay/fx-exchange/internal/models"
	"github.com/toluhikay/fx-exchange/internal/services"
	"github.com/toluhikay/fx-exchange/pkg/jwt"
	"github.com/toluhikay/fx-exchange/pkg/utils"
)

type Handler struct {
	svc     *services.Service
	userSvc *services.UserServiceImpl
}

func NewHandler(svc *services.Service, userSvc *services.UserServiceImpl) *Handler {
	return &Handler{svc: svc, userSvc: userSvc}
}

func (h *Handler) CreateWallet(w http.ResponseWriter, r *http.Request) {

	userClaims := r.Context().Value("user_claims").(*jwt.JwtClaims)

	_, err := h.svc.CreateWallet(r.Context(), userClaims.Email, userClaims.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse := utils.JSONResponse{
		Message: "wallet created successfully",
		Error:   false,
		Data:    nil,
	}
	// h.logAudit(r, walletID)
	utils.WriteJson(w, http.StatusAccepted, jsonResponse)
}

func (h *Handler) GetWallet(w http.ResponseWriter, r *http.Request) {

	userClaims := r.Context().Value("user_claims").(*jwt.JwtClaims)
	wallet, err := h.svc.GetWalletByUserId(r.Context(), userClaims.ID)
	if err != nil {
		if strings.Contains(err.Error(), "no rows in result set") {
			jsonResponse := utils.JSONResponse{
				Message: "success",
				Error:   false,
				Data:    nil,
			}
			// h.logAudit(r, walletID)
			utils.WriteJson(w, http.StatusAccepted, jsonResponse)
			return
		}
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse := utils.JSONResponse{
		Message: "success",
		Error:   false,
		Data:    wallet,
	}
	// h.logAudit(r, walletID)
	utils.WriteJson(w, http.StatusAccepted, jsonResponse)
}

func (h *Handler) Deposit(w http.ResponseWriter, r *http.Request) {
	userClaims := r.Context().Value("user_claims").(*jwt.JwtClaims)
	wallet, err := h.svc.GetWalletByUserId(r.Context(), userClaims.ID)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	walletID := wallet.ID
	var req models.DepositRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	err = h.svc.Deposit(r.Context(), walletID, req.Currency, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonResponse := utils.JSONResponse{
		Message: "deposite successful",
		Error:   false,
		Data:    nil,
	}
	h.logAudit(r, walletID)
	utils.WriteJson(w, http.StatusAccepted, jsonResponse)
}

func (h *Handler) Swap(w http.ResponseWriter, r *http.Request) {
	userClaims := r.Context().Value("user_claims").(*jwt.JwtClaims)
	wallet, err := h.svc.GetWalletByUserId(r.Context(), userClaims.ID)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	walletID := wallet.ID
	var req models.SwapRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	convertedAmount, rate, err := h.svc.Swap(r.Context(), walletID, req.FromCurrency, req.ToCurrency, req.Amount)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// h.logAudit(r, walletID)

	data := map[string]any{
		"converted_amount": convertedAmount,
		"rate":             rate,
	}

	jsonResponse := utils.JSONResponse{
		Error:   false,
		Data:    data,
		Message: "success",
	}

	utils.WriteJson(w, http.StatusAccepted, jsonResponse)
}

func (h *Handler) Transfer(w http.ResponseWriter, r *http.Request) {
	userClaims := r.Context().Value("user_claims").(*jwt.JwtClaims)
	wallet, err := h.svc.GetWalletByUserId(r.Context(), userClaims.ID)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	walletID := wallet.ID
	var req models.TransferRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	convertedAmount, rate, err := h.svc.Transfer(r.Context(), walletID, req.ReceiverID, req.Currency, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// h.logAudit(r, walletID)

	data := map[string]any{
		"converted_amount": convertedAmount,
		"rate":             rate,
	}

	jsonResponse := utils.JSONResponse{
		Error:   false,
		Data:    data,
		Message: "success",
	}

	utils.WriteJson(w, http.StatusAccepted, jsonResponse)
}

func (h *Handler) GetTransactionHistory(w http.ResponseWriter, r *http.Request) {
	userClaims := r.Context().Value("user_claims").(*jwt.JwtClaims)
	wallet, err := h.svc.GetWalletByUserId(r.Context(), userClaims.ID)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	walletID := wallet.ID
	transactions, err := h.svc.GetTransactionHistory(r.Context(), walletID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// h.logAudit(r, walletID)

	jsonResponse := utils.JSONResponse{
		Error:   false,
		Data:    transactions,
		Message: "success",
	}

	utils.WriteJson(w, http.StatusAccepted, jsonResponse)
}

func (h *Handler) logAudit(r *http.Request, walletID string) {
	// Simplified audit logging
	// Add real implementation with IP, device info, etc.
}

func (h *Handler) GetBalances(w http.ResponseWriter, r *http.Request) {
	userClaims := r.Context().Value("user_claims").(*jwt.JwtClaims)
	wallet, err := h.svc.GetWalletByUserId(r.Context(), userClaims.ID)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	walletID := wallet.ID
	balances, err := h.svc.GetBalancesWithUSD(r.Context(), walletID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// h.logAudit(r, walletID, "get_balances", "")

	jsonResponse := utils.JSONResponse{
		Error:   false,
		Data:    balances,
		Message: "success",
	}

	utils.WriteJson(w, http.StatusAccepted, jsonResponse)
}
