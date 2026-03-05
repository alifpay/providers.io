package api

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"github.com/alifpay/providers.io/internal/domain"
	"github.com/alifpay/providers.io/internal/repository"
	wf "github.com/alifpay/providers.io/internal/workflow"
	"github.com/govalues/decimal"
)

type Handler struct {
	PaymentRepo    repository.PaymentRepository
	ProviderRepo   repository.ProviderRepository
	TemporalClient client.Client
}

type CreatePaymentRequest struct {
	ProviderID int             `json:"provider_id"`
	AgentID    string          `json:"agent_id"`
	RefID      string          `json:"ref_id"`
	Amount     decimal.Decimal `json:"amount"`
	PayDate    string          `json:"pay_date"` // format: "2006-01-02", defaults to today if empty
}

type CreatePaymentResponse struct {
	PaymentID string `json:"payment_id"`
	Status    string `json:"status"`
}

// POST /payments
func (h *Handler) CreatePayment(w http.ResponseWriter, r *http.Request) {
	var req CreatePaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ProviderID == 0 || req.AgentID == "" || req.RefID == "" || !req.Amount.IsPos() {
		http.Error(w, "provider_id, agent_id, ref_id and amount are required", http.StatusBadRequest)
		return
	}

	prov, _, err := h.ProviderRepo.GetByID(r.Context(), req.ProviderID)
	if err != nil {
		log.Println(err)
		http.Error(w, "provider not found", http.StatusNotFound)
		return
	}
	if !prov.Active {
		http.Error(w, "provider is inactive", http.StatusUnprocessableEntity)
		return
	}
	if req.Amount.Cmp(prov.MinAmount) < 0 || req.Amount.Cmp(prov.MaxAmount) > 0 {
		http.Error(w, fmt.Sprintf("amount must be between %s and %s", prov.MinAmount, prov.MaxAmount), http.StatusUnprocessableEntity)
		return
	}

	payDate := time.Now().UTC().Truncate(24 * time.Hour)
	if req.PayDate != "" {
		payDate, err = time.Parse("2006-01-02", req.PayDate)
		if err != nil {
			http.Error(w, "invalid pay_date, expected format: 2006-01-02", http.StatusBadRequest)
			return
		}
	}

	payment := &domain.Payment{
		ProviderID: req.ProviderID,
		AgentID:    req.AgentID,
		RefID:      req.RefID,
		Amount:     req.Amount,
		Status:     domain.StatusPending,
		PayDate:    payDate,
	}

	paymentID, err := h.PaymentRepo.Insert(r.Context(), payment)
	if err != nil {
		slog.Error("insert payment", "err", err)
		http.Error(w, "failed to create payment", http.StatusInternalServerError)
		return
	}
	if paymentID == "" {
		// ON CONFLICT DO NOTHING — duplicate request
		http.Error(w, "duplicate payment", http.StatusConflict)
		return
	}

	workflowInput := wf.PaymentWorkflowInput{
		PaymentID:  paymentID,
		ProviderID: req.ProviderID,
		AgentID:    req.AgentID,
		RefID:      req.RefID,
		Amount:     req.Amount,
		Gate:       prov.Gate,
	}

	_, err = h.TemporalClient.ExecuteWorkflow(r.Context(),
		client.StartWorkflowOptions{
			ID:                    paymentID,
			TaskQueue:             wf.TaskQueue,
			WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_REJECT_DUPLICATE,
		},
		wf.PaymentWorkflow,
		workflowInput,
	)
	if err != nil {
		if temporal.IsWorkflowExecutionAlreadyStartedError(err) {
			http.Error(w, "duplicate payment", http.StatusConflict)
			return
		}
		slog.Error("start workflow", "err", err)
		http.Error(w, "payment accepted but workflow failed to start", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(CreatePaymentResponse{
		PaymentID: paymentID,
		Status:    "pending",
	})
}

// GET /payments/{id}
func (h *Handler) GetPayment(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	payment, err := h.PaymentRepo.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "payment not found", http.StatusNotFound)
		return
	}

	resp := map[string]any{
		"id":          payment.ID,
		"provider_id": payment.ProviderID,
		"agent_id":    payment.AgentID,
		"ref_id":      payment.RefID,
		"amount":      payment.Amount.String(),
		"fee":         payment.Fee.String(),
		"status":      strconv.Itoa(int(payment.Status)),
		"error_code":  strconv.Itoa(int(payment.ErrorCode)),
		"created_at":  payment.CreatedAt,
		"updated_at":  payment.UpdatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
