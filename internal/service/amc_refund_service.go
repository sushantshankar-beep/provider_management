package service

import (
	"context"
	"fmt"
	"log"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AMCRefundService struct {
	refund   *repository.AMCRefundRepo
	purchase *repository.AMCPurchaseRepo
	user     *repository.UserRepo
	plan     *repository.AMCPlanRepo
	payu     *PayUService
}

func NewAMCRefundService(
	r *repository.AMCRefundRepo,
	p *repository.AMCPurchaseRepo,
	u *repository.UserRepo,
	pl *repository.AMCPlanRepo,
	payu *PayUService,
) *AMCRefundService {
	return &AMCRefundService{
		refund:   r,
		purchase: p,
		user:     u,
		plan:     pl,
		payu:     payu,
	}
}

type AMCRefundListResponse struct {
	RequestID   string    `json:"requestId"`
	Date        time.Time `json:"date"`
	User        string    `json:"user"`
	Email       string    `json:"email"`
	Plan        string    `json:"plan"`
	Reason      string    `json:"reason"`
	ServiceUsed int       `json:"serviceUsed"`
	Amount      float64   `json:"amount"`
	GST         float64   `json:"gst"`
	Discount    float64   `json:"discount"`
	Refund      float64   `json:"refund"`
	Status      string    `json:"status"`
}

type RefundDetailResponse struct {
	RequestID          string              `json:"requestId"`
	OrderDetails       OrderDetails        `json:"orderDetails"`
	CancellationReason string              `json:"cancellationReason"`
	RefundPayment      RefundPayment       `json:"refundPayment"`
	Status             string              `json:"status"`
	Timeline           []domain.Timeline   `json:"timeline"`
	RejectionReason    string              `json:"rejectionReason,omitempty"`
	RefundTransactionID string             `json:"refundTransactionId,omitempty"`
}

type OrderDetails struct {
	User         string    `json:"user"`
	Email        string    `json:"email"`
	Plan         string    `json:"plan"`
	PurchaseDate time.Time `json:"purchaseDate"`
	AmountPaid   float64   `json:"amountPaid"`
	ServiceUsed  int       `json:"serviceUsed"`
	IsActivated  bool      `json:"isActivated"`
}

type RefundPayment struct {
	BasePrice      float64 `json:"basePrice"`
	TotalAmount    float64 `json:"totalAmount"`
	GSTAmount      float64 `json:"gstAmount"`
	DiscountAmount float64 `json:"discountAmount"`
	RefundAmount   float64 `json:"refundAmount"`
}

type ApproveRefundResponse struct {
	Status              string            `json:"status"`
	RefundTransactionID string            `json:"refundTransactionId"`
	PayURequestID       string            `json:"payuRequestId"`
	Timeline            []domain.Timeline `json:"timeline"`
}

type RejectRefundResponse struct {
	Status          string            `json:"status"`
	RejectionReason string            `json:"rejectionReason"`
	Timeline        []domain.Timeline `json:"timeline"`
}

type CheckStatusResponse struct {
	Status       string            `json:"status"`
	Timeline     []domain.Timeline `json:"timeline"`
	RefundDetails RefundDetails    `json:"refundDetails"`
}

type RefundDetails struct {
	Amount       float64 `json:"amount"`
	BankRefNum   string  `json:"bankRefNum,omitempty"`
	SettlementID string  `json:"settlementId,omitempty"`
	RefundMode   string  `json:"refundMode,omitempty"`
}

func (s *AMCRefundService) ListRefundRequests(
	ctx context.Context,
	pageStr, limitStr, search, status, sortBy, sortOrder string,
) ([]AMCRefundListResponse, int64, error) {
	page, _ := strconv.ParseInt(pageStr, 10, 64)
	limit, _ := strconv.ParseInt(limitStr, 10, 64)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	skip := (page - 1) * limit

	refunds, total, err := s.refund.FindWithFilter(ctx, skip, limit, search, status, sortBy, sortOrder)
	if err != nil {
		return nil, 0, err
	}

	var result []AMCRefundListResponse
    log.Println("edjcbhjwsbhje")
	for _, ref := range refunds {
		resp := AMCRefundListResponse{
			RequestID: ref.ID.Hex(),
			Date:      ref.CreatedAt,
			Reason:    ref.Reason,
			Refund:    ref.RefundAmount,
			Status:    ref.Status,
		}
		log.Println("edjcbhjwsbhje2222")
		if ref.UserID != primitive.NilObjectID {
			if user, err := s.user.FindByID(ctx, ref.UserID.Hex()); err == nil {
				log.Println("userrr",user)
				resp.User = user.Name
				resp.Email = user.Email
			}
		}
        log.Println("djkebkjbf",ref.AMCPurchaseID)
		if ref.AMCPurchaseID != primitive.NilObjectID {
			if purchase, err := s.purchase.FindByID(ctx, ref.AMCPurchaseID.Hex()); err == nil {
				log.Println("purchaseeee",purchase)
				resp.Amount = purchase.PlanPrice
                log.Println(purchase.PlanName)
				serviceUsed := 0
				for _, sd := range purchase.ServiceDetails {
					if count, err := strconv.Atoi(sd.Count); err == nil {
						serviceUsed += count
					}
				}
				resp.ServiceUsed = serviceUsed

				if purchase.PlanID != primitive.NilObjectID {
					if plan, err := s.plan.FindByID(ctx, purchase.PlanID.Hex()); err == nil {
						log.Println("planIdddd",plan)
						resp.Plan = plan.PlanName
						resp.GST = plan.PlanGSTAmount
						resp.Discount = plan.PlanDiscountAmount
					}
				}
			}
		}

		result = append(result, resp)
	}

	return result, total, nil
}

func (s *AMCRefundService) GetRefundDetails(ctx context.Context, id string) (*RefundDetailResponse, error) {
	refund, err := s.refund.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Refund request not found")
	}

	resp := &RefundDetailResponse{
		RequestID:           refund.ID.Hex(),
		CancellationReason:  refund.Reason,
		Status:              refund.Status,
		Timeline:            refund.Timeline,
		RejectionReason:     refund.RejectionReason,
		RefundTransactionID: refund.RefundTransactionID,
	}

	if refund.UserID != primitive.NilObjectID {
		if user, err := s.user.FindByID(ctx, refund.UserID.Hex()); err == nil {
			resp.OrderDetails.User = user.Name
			resp.OrderDetails.Email = user.Email
		}
	}

	if refund.AMCPurchaseID != primitive.NilObjectID {
		if purchase, err := s.purchase.FindByID(ctx, refund.AMCPurchaseID.Hex()); err == nil {
			resp.OrderDetails.PurchaseDate = purchase.CreatedAt
			resp.OrderDetails.AmountPaid = purchase.PlanPrice
			resp.OrderDetails.IsActivated = purchase.PlanStatus == "active"

			serviceUsed := 0
			for _, sd := range purchase.ServiceDetails {
				if count, err := strconv.Atoi(sd.Count); err == nil {
					serviceUsed += count
				}
			}
			resp.OrderDetails.ServiceUsed = serviceUsed

			resp.RefundPayment.TotalAmount = purchase.PlanPrice

			if purchase.PlanID != primitive.NilObjectID {
				if plan, err := s.plan.FindByID(ctx, purchase.PlanID.Hex()); err == nil {
					resp.OrderDetails.Plan = plan.PlanName
					resp.RefundPayment.BasePrice = plan.PlanBasePrice
					resp.RefundPayment.GSTAmount = plan.PlanGSTAmount
					resp.RefundPayment.DiscountAmount = plan.PlanDiscountAmount
				}
			}
		}
	}

	resp.RefundPayment.RefundAmount = refund.RefundAmount

	return resp, nil
}

func (s *AMCRefundService) ApproveRefund(ctx context.Context, id, adminNote, adminID string) (*ApproveRefundResponse, error) {
	refund, err := s.refund.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Refund request not found")
	}

	if refund.Status == "cancelled" {
		return nil, fmt.Errorf("Refund request is already cancelled and cannot be updated")
	}

	if refund.Status == "approved" {
		return nil, fmt.Errorf("Refund request is already approved")
	}

	if refund.Status == "rejected_admin" || refund.Status == "rejected_payu" {
		return nil, fmt.Errorf("Refund request is already rejected")
	}

	purchase, err := s.purchase.FindByID(ctx, refund.AMCPurchaseID.Hex())
	if err != nil {
		return nil, fmt.Errorf("AMC purchase not found")
	}

	payuResult, err := s.payu.InitiateRefund(ctx, refund, purchase)
	if err != nil {
		return nil, err
	}

	adminObjID, _ := primitive.ObjectIDFromHex(adminID)
	now := time.Now()

	timeline := domain.Timeline{
		Status:      "under_process",
		Timestamp:   now,
		Description: "Refund initiated with PayU, awaiting confirmation.",
		Note:        adminNote,
	}

	err = s.refund.UpdateStatus(ctx, id, "under_process", payuResult.RequestID, payuResult.RefundTransactionID, payuResult.PayUResponse, adminObjID, timeline)
	if err != nil {
		return nil, err
	}

	err = s.purchase.UpdateRefundStatus(ctx, refund.AMCPurchaseID.Hex(), "under_process")
	if err != nil {
		return nil, err
	}

	refund, _ = s.refund.FindByID(ctx, id)

	return &ApproveRefundResponse{
		Status:              refund.Status,
		RefundTransactionID: refund.RefundTransactionID,
		PayURequestID:       refund.PayURequestID,
		Timeline:            refund.Timeline,
	}, nil
}

func (s *AMCRefundService) RejectRefund(ctx context.Context, id, adminNote, adminID string) (*RejectRefundResponse, error) {
	refund, err := s.refund.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Refund request not found")
	}

	if refund.Status == "cancelled" {
		return nil, fmt.Errorf("Refund request is already cancelled and cannot be updated")
	}

	if refund.Status == "approved" {
		return nil, fmt.Errorf("Refund request is already approved and cannot be rejected")
	}

	if refund.Status == "rejected_admin" || refund.Status == "rejected_payu" {
		return nil, fmt.Errorf("Refund request is already rejected")
	}

	adminObjID, _ := primitive.ObjectIDFromHex(adminID)
	now := time.Now()

	timeline := domain.Timeline{
		Status:      "rejected_admin",
		Timestamp:   now,
		Description: "Refund request has been rejected by admin.",
		Note:        adminNote,
	}

	err = s.refund.UpdateStatusRejected(ctx, id, "rejected_admin", adminNote, adminObjID, timeline)
	if err != nil {
		return nil, err
	}

	err = s.purchase.UpdateRefundStatus(ctx, refund.AMCPurchaseID.Hex(), "rejected_admin")
	if err != nil {
		return nil, err
	}

	refund, _ = s.refund.FindByID(ctx, id)

	return &RejectRefundResponse{
		Status:          refund.Status,
		RejectionReason: refund.RejectionReason,
		Timeline:        refund.Timeline,
	}, nil
}

func (s *AMCRefundService) CheckRefundStatus(ctx context.Context, id string) (*CheckStatusResponse, error) {
	refund, err := s.refund.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Refund request not found")
	}

	if refund.Status != "under_process" {
		return &CheckStatusResponse{
			Status:   refund.Status,
			Timeline: refund.Timeline,
			RefundDetails: RefundDetails{
				Amount:       refund.RefundAmount,
				BankRefNum:   refund.BankRefNum,
				SettlementID: refund.SettlementID,
				RefundMode:   refund.RefundMode,
			},
		}, nil
	}

	if refund.PayURequestID == "" {
		return nil, fmt.Errorf("PayU request ID not found")
	}

	statusResult, err := s.payu.CheckRefundStatus(ctx, refund.PayURequestID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	if statusResult.RefundStatus == "success" {
		timeline := domain.Timeline{
			Status:      "approved",
			Timestamp:   now,
			Description: "Refund completed successfully",
			Note:        fmt.Sprintf("Amount refunded: ₹%.2f via %s", statusResult.Amount, statusResult.Mode),
		}

		err = s.refund.UpdateStatusApproved(
			ctx,
			id,
			statusResult.BankRefNum,
			statusResult.SettlementID,
			statusResult.Mode,
			statusResult.BankArn,
			statusResult.RawResponse,
			timeline,
		)
		
		if err != nil {
			return nil, err
		}

		err = s.purchase.UpdateRefundStatusAndPlanStatus(ctx, refund.AMCPurchaseID.Hex(), "approved", "cancelled")
		if err != nil {
			return nil, err
		}
	} else if strings.Contains(strings.ToLower(statusResult.RefundStatus), "fail") || 
		strings.Contains(strings.ToLower(statusResult.RefundStatus), "error") || 
		strings.Contains(strings.ToLower(statusResult.RefundStatus), "reject") {
		
		timeline := domain.Timeline{
			Status:      "rejected_payu",
			Timestamp:   now,
			Description: "Refund failed",
			Note:        statusResult.ErrorMsg,
		}

		err = s.refund.UpdateStatusRejectedPayU(ctx, id, timeline)
		if err != nil {
			return nil, err
		}

		err = s.purchase.UpdateRefundStatus(ctx, refund.AMCPurchaseID.Hex(), "rejected_payu")
		if err != nil {
			return nil, err
		}
	} else {
		err = s.refund.UpdateLastStatusCheck(ctx, id, statusResult.RawResponse)
		if err != nil {
			return nil, err
		}
	}

	refund, _ = s.refund.FindByID(ctx, id)

	return &CheckStatusResponse{
		Status:   refund.Status,
		Timeline: refund.Timeline,
		RefundDetails: RefundDetails{
			Amount:       refund.RefundAmount,
			BankRefNum:   refund.BankRefNum,
			SettlementID: refund.SettlementID,
			RefundMode:   refund.RefundMode,
		},
	}, nil
}


func (s *AMCRefundService) GetRefundStats(ctx context.Context) (*repository.RefundStats, error) {
	return s.refund.GetStats(ctx)
}
