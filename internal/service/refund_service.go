package service

import (
	"context"
	"fmt"
	"log"
	"time"
	"provider_management/internal/domain"
	"provider_management/internal/dto"
	"provider_management/internal/repository"
	"provider_management/internal/utils"
	"go.mongodb.org/mongo-driver/bson"
)

type RefundService struct {
	refundRepo      *repository.RefundRepository
	transactionRepo *repository.TransactionRepo
	userRepo        *repository.UserRepo
	complaintRepo   *repository.ComplaintRepository
	acceptedService *repository.AcceptedServiceRepo
	payuService     *PayUService
}

func NewRefundService(
	refundRepo *repository.RefundRepository,
	transactionRepo *repository.TransactionRepo,
	userRepo *repository.UserRepo,
	complaintRepo *repository.ComplaintRepository,
	acceptedService *repository.AcceptedServiceRepo,
	payuService     *PayUService,
) *RefundService {
	return &RefundService{
		refundRepo:      refundRepo,
		transactionRepo: transactionRepo,
		userRepo:        userRepo,
		complaintRepo:   complaintRepo,
		acceptedService: acceptedService,
		payuService: payuService,
	}
}

func (s *RefundService) ProcessRefund(ctx context.Context, req dto.RefundRequest) error {

	now := time.Now()

	transaction, err := s.transactionRepo.FindByTxnID(ctx, req.TxnID)
	if err != nil {
		return fmt.Errorf("transaction not found with txnid %s: %w", req.TxnID, err)
	}
	
	gstAmount := req.Amount * 0.18

	refund := &domain.Refund{
		UserID:        req.UserID,
		TxnID:         transaction.TxnID,
		ServiceID:     req.BookingID,      
	    ComplaintID:   req.ComplaintID,   
		Reason:        req.Reason,
		Amount:        utils.RoundTo2(req.Amount),
		GST:           utils.RoundTo2(gstAmount),
		NetRefund:     utils.RoundTo2(req.Amount),
		Mode:          transaction.Method,
		Status:        domain.RefundStatusPending,
		Timeline: domain.RefundTimeline{
			Initiated: now,
		},
	}

	if err := s.refundRepo.Create(ctx, refund); err != nil {
		return fmt.Errorf("failed to create refund: %w", err)
	}

	if err := s.transactionRepo.UpdateRefundID(ctx, transaction.ID, refund.ID); err != nil {
		log.Printf("ProcessRefund - Failed to update transaction with refund ID: %v", err)
	}

	return nil
}

func (s *RefundService) GetAllRefunds(
	ctx context.Context,
	filter domain.RefundFilter,
) (*dto.RefundListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 10
	}

	skip := (filter.Page - 1) * filter.Limit

	refunds, total, err := s.refundRepo.FindAll(ctx, filter, skip, filter.Limit)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch refunds: %w", err)
	}

	userIDs := make([]string, 0, len(refunds))
	serviceIDs := make([]string, 0, len(refunds))

	for _, r := range refunds {
		userIDs = append(userIDs, r.UserID)
		if r.ServiceID != "" {
			serviceIDs = append(serviceIDs, r.ServiceID)
		}
	}

	users, err := s.userRepo.FindByIDs(ctx, userIDs)
	if err != nil {
		log.Printf("Failed to fetch users: %v", err)
	}

	userMap := make(map[string]*domain.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	acceptedServices := make(map[string]*domain.AcceptedService)
	for _, serviceID := range serviceIDs {
		service, err := s.acceptedService.FindByServiceRequestID(ctx, serviceID)
		if err == nil && service != nil {
			acceptedServices[serviceID] = service
		}
	}

	acceptedServiceIDs := make([]string, 0, len(acceptedServices))
	for _, service := range acceptedServices {
		acceptedServiceIDs = append(acceptedServiceIDs, service.ID.Hex())
	}

	complaints := make(map[string]*domain.Complaint)
	for _, acceptedServiceID := range acceptedServiceIDs {
		complaint, err := s.complaintRepo.FindComplaintByAcceptedServiceId(ctx, acceptedServiceID)
		if err == nil && complaint != nil {
			complaints[acceptedServiceID] = complaint
		}
	}

	items := make([]dto.RefundListItemDTO, 0, len(refunds))

	for _, refund := range refunds {
		userCode := refund.UserID
		
		if u, ok := userMap[refund.UserID]; ok {
			userCode = u.UserCode
		}

		var serviceNumber string
		var complaintNumber string

		if service, ok := acceptedServices[refund.ServiceID]; ok {
			serviceNumber = service.ServiceNumber
			if complaint, exists := complaints[service.ID.Hex()]; exists {
				complaintNumber = complaint.ComplaintNumber
			}
		}

		items = append(items, dto.RefundListItemDTO{
			ID:              refund.ID,
			RefundID:        refund.RefundID,
			UserCode:        userCode,
			NetRefund: refund.NetRefund,
			ServiceNumber:   serviceNumber,
			ComplaintNumber: complaintNumber,
			TxnID:   refund.TxnID,
			GST:             utils.RoundTo2(refund.GST),
			Mode:            refund.Mode,
			Amount:          utils.RoundTo2(refund.Amount),
			Status:          refund.Status,
			Reason:          refund.Reason,
			SubmittedAt:     refund.CreatedAt,
		})
	}

	return &dto.RefundListResponse{
		Refunds:    items,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: (total + int64(filter.Limit) - 1) / int64(filter.Limit),
	}, nil
}

func (s *RefundService) GetRefundByID(
	ctx context.Context,
	refundID string,
) (*dto.RefundDetailDTO, error) {
	r, err := s.refundRepo.FindByID(ctx, refundID)
	if err != nil {
		return nil, fmt.Errorf("refund not found: %w", err)
	}

	user, err := s.userRepo.FindByID(ctx, r.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found for refund: %w", err)
	}

	var serviceNumber string
	var complaintNumber string
	var complaintId string

	if r.ServiceID != "" {
		service, err := s.acceptedService.FindByServiceRequestID(ctx, r.ServiceID)
		if err == nil && service != nil {
			serviceNumber = service.ServiceNumber

			complaint, err := s.complaintRepo.FindComplaintByAcceptedServiceId(ctx, service.ID.Hex())
			if err == nil && complaint != nil {
				complaintNumber = complaint.ComplaintNumber
				complaintId = complaint.ID
			}
		}
	}

	return &dto.RefundDetailDTO{
		ID:              r.ID,
		RefundID:        r.RefundID,
		UserCode:        user.UserCode,
		Amount:          utils.RoundTo2(r.Amount),
		ServiceNumber:   serviceNumber,
		NetRefund: r.NetRefund,
		ComplaintNumber: complaintNumber,
		ComplaintID: complaintId,
		TransactionID:   r.TxnID,
		GST:             utils.RoundTo2(r.GST),
		Mode:            r.Mode,
		Reason:          r.Reason,
		Status:          r.Status,
		Timeline: r.Timeline,
		CreatedAt:       r.CreatedAt,
	}, nil
}

func (s *RefundService) InitiateRefund(ctx context.Context, refundID string) error {

	refund, err := s.refundRepo.FindByID(ctx, refundID)
	if err != nil {
		return fmt.Errorf("refund not found: %w", err)
	}

	if refund.Status != domain.RefundStatusPending {
		return fmt.Errorf("refund cannot be initiated, current status: %s", refund.Status)
	}

	transaction, err := s.transactionRepo.FindByTxnID(ctx, refund.TxnID)
	if err != nil {
		return fmt.Errorf("transaction not found: %w", err)
	}

	if transaction.MihPayID == "" {
		return fmt.Errorf("missing PayU transaction ID")
	}

	payuRefundID := fmt.Sprintf(
		"REF%s%d",
		refund.ID.Hex()[len(refund.ID.Hex())-6:],
		time.Now().Unix()%1000000,
	)
    
	totalAmount := refund.Amount

	payuResp, err := s.payuService.InitiateRefundByParams(
		ctx,
		transaction.MihPayID,
		payuRefundID,
		totalAmount,
	)

	// ---------------- FAILED CASE ----------------
	if err != nil {
		log.Printf("PayU refund initiation failed: %v", err)

		update := bson.M{
			"$set": bson.M{
				"status":           domain.RefundStatusFailed,
				"failureReason":    err.Error(),
				"timeline.failed":  time.Now(),
			},
		}

		
		_ = s.refundRepo.Update(ctx, refund.ID, update)
		return fmt.Errorf("failed to initiate refund with PayU: %w", err)
	}

	// ---------------- SUCCESS CASE ----------------
	update := bson.M{
		"$set": bson.M{
			"status":                domain.RefundStatusUnderProcess,
			"payuRequestId":         payuResp.RequestID,
			"payuTransactionId":     payuResp.RefundTransactionID,
			"payuRefundResponse":    payuResp.PayUResponse,
			"timeline.underProcess": time.Now(),
		},
	}
	

	if err := s.refundRepo.Update(ctx, refund.ID, update); err != nil {
		return fmt.Errorf("refund initiated but failed to update record: %w", err)
	}

	return nil
}

func (s *RefundService) CheckRefundStatus(ctx context.Context, refundID string) error {
	refund, err := s.refundRepo.FindByID(ctx, refundID)
	if err != nil {
		return fmt.Errorf("refund not found: %w", err)
	}

	if refund.Status == domain.RefundStatusSuccess || refund.Status == domain.RefundStatusFailed {
		return fmt.Errorf("refund already in final state: %s", refund.Status)
	}

	if refund.PayURequestID == "" {
		return fmt.Errorf("refund not initiated with PayU yet")
	}

	statusResp, err := s.payuService.CheckRefundStatus(ctx, refund.PayURequestID)
	if err != nil {
		return fmt.Errorf("failed to check refund status: %w", err)
	}

	now := time.Now()

	setFields := bson.M{
		"payuStatusResponse":     statusResp.RawResponse,
		"timeline.statusChecked": now,
	}

	var newStatus domain.RefundStatus

	switch statusResp.RefundStatus {
	case "success", "successful", "completed":
		newStatus = domain.RefundStatusSuccess
		setFields["bankRefNum"] = statusResp.BankRefNum
		setFields["refundMode"] = statusResp.Mode
		setFields["settlementId"] = statusResp.SettlementID
		setFields["bankArn"] = statusResp.BankArn
		setFields["timeline.completed"] = now

	case "failed", "failure":
		newStatus = domain.RefundStatusFailed
		reason := "Refund failed at payment gateway"
		if statusResp.ErrorMsg != "" {
			reason = statusResp.ErrorMsg
		}
		setFields["failureReason"] = reason
		setFields["timeline.failed"] = now

	case "pending", "initiated", "processing":
		newStatus = domain.RefundStatusUnderProcess

	default:
		newStatus = refund.Status
	}

	if newStatus != refund.Status {
		setFields["status"] = newStatus
	}

	update := bson.M{
		"$set": setFields,
	}

	if err := s.refundRepo.Update(ctx, refund.ID, update); err != nil {
		return fmt.Errorf("status checked but failed to update record: %w", err)
	}

	return nil
}

