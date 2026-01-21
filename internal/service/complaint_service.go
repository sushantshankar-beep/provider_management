package service

import (
	"fmt"
	"time"
	"strconv"
	"strings"
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"provider_management/internal/dto"
	"provider_management/internal/utils"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ComplaintService struct {
	paymentPayoutRepo   *repository.PaymentPayoutRepo
	complaintRepo       *repository.ComplaintRepository
	acceptedServiceRepo *repository.AcceptedServiceRepo
	userRepo            *repository.UserRepo
	providerRepo        *repository.ProviderRepo
	refundService       *RefundService
	payoutService       *PayoutService
}

func NewComplaintService(
	paymentPayoutRepo *repository.PaymentPayoutRepo,
	complaintRepo *repository.ComplaintRepository,
	acceptedServiceRepo *repository.AcceptedServiceRepo,
	userRepo *repository.UserRepo,
	providerRepo *repository.ProviderRepo,
	refundService *RefundService,
	payoutService *PayoutService,
) *ComplaintService {
	return &ComplaintService{
		paymentPayoutRepo:   paymentPayoutRepo,
		complaintRepo:       complaintRepo,
		acceptedServiceRepo: acceptedServiceRepo,
		userRepo:            userRepo,
		providerRepo:        providerRepo,
		refundService:       refundService,
		payoutService:       payoutService,
	}
}

func (s *ComplaintService) GetComplaint(ctx context.Context, id string) (*domain.Complaint, error) {

	cleanID := strings.TrimPrefix(id, "CMP")

	if primitive.IsValidObjectID(cleanID) {
		complaint, err := s.complaintRepo.GetByID(ctx, cleanID)
		if err == nil {
			return complaint, nil
		}
	}

	if internalID, parseErr := strconv.ParseInt(cleanID, 10, 64); parseErr == nil {
		complaint, err := s.complaintRepo.GetByInternalID(ctx, internalID)
		if err == nil {
			return complaint, nil
		}

	}

	return nil, fmt.Errorf("complaint not found with ID: %s", id)
}

func (s *ComplaintService) GetAllComplaints(ctx context.Context, filters dto.ComplaintFilters, pagination dto.ComplaintPagination) (*dto.ComplaintListResponse, error) {
	if pagination.Page < 1 {
		pagination.Page = 1
	}

	if pagination.Limit < 1 {
		pagination.Limit = 10
	}

	domainFilter := dto.ComplaintFilter{
		Page:          pagination.Page,
		Limit:         pagination.Limit,
		Status:        filters.Status,
		RaisedBy:      filters.RaisedBy,
		Category:      filters.Category,
		SearchQuery:   filters.SearchQuery,
		UserID:        filters.UserID,
		ProviderID:    filters.ProviderID,
		CreatedAtFrom: filters.CreatedAtFrom,
		CreatedAtTo:   filters.CreatedAtTo,
	}

	complaints, total, stats, err := s.complaintRepo.ListComplaints(ctx, domainFilter)

	if err != nil {
		return nil, err
	}

	data := make([]dto.ComplaintListItem, 0, len(complaints))

	for _, complaint := range complaints {
		indianTime := complaint.CreatedAt.Add(5*time.Hour + 30*time.Minute)
		resp := dto.ComplaintListItem{
			ID:                complaint.ID,
			InternalID:        "CMP" + strconv.FormatInt(complaint.InternalID, 10),
			AcceptedServiceNo: "BK" + strconv.FormatInt(complaint.AcceptedServiceNo, 10),
			RaisedBy:          complaint.RaisedBy,
			Problem:           complaint.Problem,
			Status:            complaint.Status,
			CreatedAt:         indianTime.Format("2006-01-02 15:04:05"),
		}

		data = append(data, resp)
	}

	totalPages := (total + int64(pagination.Limit) - 1) / int64(pagination.Limit)

	return &dto.ComplaintListResponse{
		Data:  data,
		Stats: stats,
		Pagination: dto.PaginationMeta{
			Page:        int64(pagination.Page),
			Limit:       int64(pagination.Limit),
			TotalItems:  total,
			TotalPages:  totalPages,
			HasNext:     int64(pagination.Page) < totalPages,
			HasPrevious: pagination.Page > 1,
		},
	}, nil
}

func (s *ComplaintService) GetComplaintWithDetails(ctx context.Context, id string) (*dto.ComplaintDetailResponse, error) {
	complaint, err := s.GetComplaint(ctx, id)

	if err != nil {
		return nil, err
	}

	complaintWithDetails := &domain.ComplaintWithDetails{
		Complaint: *complaint,
	}

	var booking *domain.AcceptedService
	if complaint.AcceptedServiceID != "" {
		booking, _ = s.acceptedServiceRepo.FindByID(ctx, complaint.AcceptedServiceID)
		complaintWithDetails.BookingDetails = &domain.BookingDetails{
			ID:         booking.ID.Hex(),
			InternalID: booking.InternalID,
			BasePrice:  booking.BasePrice,
			FinalPrice: booking.FinalPrice,
		}
	}

	userID := complaint.UserID
	if userID == "" && booking != nil {
		userID = booking.UserID
	}

	if userID != "" {
		user, _ := s.userRepo.FindByID(ctx, userID)
		complaintWithDetails.UserDetails = &domain.UserDetails{
			ID:         user.ID,
			InternalID: fmt.Sprintf("VW%d", user.InternalID),
			Name:       user.Name,
			Email:      user.Email,
			Phone:      user.Phone,
		}
	}

	providerID := complaint.ProviderID
	if providerID == "" && booking != nil {
		providerID = booking.ProviderID.Hex()
	}

	if providerID != "" {
		provider, _ := s.providerRepo.FindByID(ctx, providerID)
		complaintWithDetails.ProviderDetails = &domain.ProviderDetails{
			ID:          provider.ID.Hex(),
			InternalID:  fmt.Sprintf("PRO%d", provider.InternalID),
			Name:        provider.Name,
			Email:       provider.Email,
			Phone:       provider.Phone,
			CompanyName: provider.CompanyName,
		}
	}

	resp := &dto.ComplaintDetailResponse{
		ID:               complaintWithDetails.ID,
		ComplaintID:      "CMP" + strconv.FormatInt(complaintWithDetails.InternalID, 10),
		BookingID:        complaintWithDetails.AcceptedServiceID,
		BookingNo:        "BK" + strconv.FormatInt(complaintWithDetails.AcceptedServiceNo, 10),
		UserID:           complaintWithDetails.UserID,
		ProviderID:       complaintWithDetails.ProviderID,
		RaisedBy:         complaintWithDetails.RaisedBy,
		Problem:          complaintWithDetails.Problem,
		Photos:           complaintWithDetails.Photos,
		Status:           complaintWithDetails.Status,
		Timeline:         complaintWithDetails.Timeline,
		Assessment:       complaintWithDetails.Assessment,
		Notes:            complaintWithDetails.Notes,
		ActionsTriggered: complaintWithDetails.ActionsTriggered,
		CreatedAt:        complaintWithDetails.CreatedAt,
		UpdatedAt:        complaintWithDetails.UpdatedAt,
		UpdatedByAdmin:   complaintWithDetails.UpdatedByAdmin,
		AdminUpdatedAt:   complaintWithDetails.AdminUpdatedAt,
		PaymentTracking:  complaintWithDetails.PaymentTracking,
		UserDetails:      complaintWithDetails.UserDetails,
		ProviderDetails:  complaintWithDetails.ProviderDetails,
		BookingDetails:   complaintWithDetails.BookingDetails,
	}

	return resp, nil
}

func (s *ComplaintService) AssessComplaint(ctx context.Context, complaintID string, req dto.AssessComplaintRequest) error {
	complaint, err := s.GetComplaint(ctx, complaintID)

	if err != nil {
		return fmt.Errorf("failed to get complaint: %w", err)
	}

	if complaint.Status != domain.ComplaintStatusInReview {
		return fmt.Errorf("complaint must be in in_review status to save assessment, current status: %s", complaint.Status)
	}

	acceptedService, err := s.acceptedServiceRepo.FindByID(ctx, complaint.AcceptedServiceID)

	if err != nil {
		return fmt.Errorf("failed to get accepted service: %w", err)
	}

	if acceptedService.OrderID == "" {
		return fmt.Errorf("no transaction associated with this booking")
	}

	if err := s.validateComplaintAssessmentAmounts(req, acceptedService.BasePrice); err != nil {
		return err
	}

	if err := s.saveAssessmentAndResolve(ctx, complaint.ID, req); err != nil {
		return err
	}

	return s.processPaymentActions(ctx, complaint, acceptedService, req)

}

func (s *ComplaintService) validateComplaintAssessmentAmounts(req dto.AssessComplaintRequest, originalAmount float64) error {

	if originalAmount <= 0 {
		return fmt.Errorf("invalid original booking amount: %.2f", originalAmount)
	}

	if req.RefundToUser == domain.RefundTypeFull {
		req.RefundAmount = originalAmount
	}

	if req.PayoutToProvider == domain.PayoutTypeFull {
		req.PayoutAmount = originalAmount
	}

	if req.RefundToUser == domain.RefundTypePartial {
		if req.RefundAmount <= 0 || req.RefundAmount > originalAmount {
			return fmt.Errorf("invalid refund amount: %.2f", req.RefundAmount)
		}
	}

	if req.PayoutToProvider == domain.PayoutTypePartial {
		if req.PayoutAmount <= 0 || req.PayoutAmount > originalAmount {
			return fmt.Errorf("invalid payout amount: %.2f", req.PayoutAmount)
		}
	}

	if req.RefundAmount+req.PayoutAmount > originalAmount {
		return fmt.Errorf("refund (%.2f) + payout (%.2f) exceeds payable booking amount %.2f", req.RefundAmount, req.PayoutAmount, originalAmount)
	}

	return nil
}

func (s *ComplaintService) saveAssessmentAndResolve(ctx context.Context, complaintID string, req dto.AssessComplaintRequest) error {

	assessment := domain.ComplaintAssessment{
		FaultParty:       req.FaultParty,
		RefundToUser:     req.RefundToUser,
		RefundAmount:     req.RefundAmount,
		PayoutToProvider: req.PayoutToProvider,
		PayoutAmount:     req.PayoutAmount,
		Remarks:          req.Remarks,
		AssessedBy:       req.AssessedBy,
	}

	if err := s.complaintRepo.SaveAssessment(ctx, complaintID, assessment); err != nil {
		return fmt.Errorf("failed to save assessment: %w", err)
	}

	if err := s.complaintRepo.UpdateStatus(ctx, complaintID, domain.ComplaintStatusResolved); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

func (s *ComplaintService) processPaymentActions(ctx context.Context, complaint *domain.Complaint, acceptedService *domain.AcceptedService, req dto.AssessComplaintRequest) error {

	paymentTracking := &domain.PaymentActionTracking{
		RefundStatus: domain.PaymentActionNA,
		PayoutStatus: domain.PaymentActionNA,
	}

	actions := []string{}

	if req.RefundToUser != domain.RefundTypeNone && req.RefundAmount > 0 {
		refundAction := s.processRefund(ctx, complaint, acceptedService.OrderID, req)
		if refundAction != "" {
			paymentTracking.RefundStatus = domain.PaymentActionPending
			actions = append(actions, refundAction)
		}
	}

	if !acceptedService.ProviderID.IsZero() {
		payoutActions := s.processPayout(ctx, complaint, acceptedService, req)
		if len(payoutActions) > 0 {
			paymentTracking.PayoutStatus = domain.PaymentActionPending
			actions = append(actions, payoutActions...)
		}
	}

	if len(actions) == 0 {
		actions = append(actions, "No financial actions triggered")
	}

	return s.complaintRepo.Update(ctx, complaint.ID, bson.M{
		"actionsTriggered": actions,
		"paymentTracking":  paymentTracking,
	})
}

func (s *ComplaintService) processRefund(ctx context.Context, complaint *domain.Complaint, txnID string, req dto.AssessComplaintRequest) string {

	userObjID, err := primitive.ObjectIDFromHex(complaint.UserID)

	if err != nil {
		fmt.Printf("Warning: Invalid UserID format: %s", complaint.UserID)
		return ""
	}

	refundReason := req.Remarks

	if req.RefundToUser == domain.RefundTypeFull {
		refundReason = refundReason
	} else {
		refundReason = refundReason
	}

	if err := s.refundService.ProcessRefund(ctx, dto.RefundRequest{
		UserID:              userObjID.Hex(),
		TxnID:               txnID,
		BookingID:           complaint.AcceptedServiceID,
		BookingInternalID:   complaint.AcceptedServiceNo,
		ComplaintID:         complaint.ID,
		ComplaintInternalID: complaint.InternalID,
		Amount:              req.RefundAmount,
		Reason:              refundReason,
	}); err != nil {
		fmt.Printf("Warning: Failed to process refund: %v", err)
		return ""
	}

	actionMsg := fmt.Sprintf("Refund of %.2f sent to Refund Management module", req.RefundAmount)
	if req.RefundToUser == domain.RefundTypeFull {
		actionMsg = "Full " + actionMsg
	} else {
		actionMsg = "Partial " + actionMsg
	}

	return actionMsg
}

func (s *ComplaintService) processPayout(ctx context.Context, complaint *domain.Complaint, acceptedService *domain.AcceptedService, req dto.AssessComplaintRequest) []string {

	providerID := acceptedService.ProviderID.Hex()
	originalAmount := acceptedService.BasePrice
	isNoPayout := req.PayoutToProvider == domain.PayoutTypeNone || req.PayoutToProvider == "No Payout"
	isSettled := acceptedService.SettlementStatus == domain.SettleStatusPending || acceptedService.SettlementStatus == domain.SettleStatusSettled

	if isNoPayout && isSettled {
		return s.handleNoPayoutSettled(ctx, complaint, acceptedService, providerID, originalAmount)
	}

	if isNoPayout && !isSettled {
		return s.handleNoPayoutNotSettled(ctx, complaint, acceptedService, providerID, originalAmount)
	}

	if !isNoPayout && req.PayoutAmount > 0 {
		return s.handlePartialPayout(ctx, complaint, acceptedService, providerID, originalAmount, req.PayoutAmount, isSettled)
	}

	return []string{}
}

func (s *ComplaintService) handleNoPayoutSettled(ctx context.Context, complaint *domain.Complaint, acceptedService *domain.AcceptedService, providerID string, originalAmount float64) []string {

	commission := originalAmount * 20.0 / 100
	gst := (originalAmount - commission) * 18.0 / 100
	netPayable := originalAmount - commission - gst

	err := s.payoutService.ProcessPayout(ctx, dto.PayoutRequest{
		ProviderID:          providerID,
		BookingID:           complaint.AcceptedServiceID,
		Amount:              originalAmount,
		PartialAmount:       0,
		CancelPayout:        false,
		CreateDeduction:     true,
		Reason:              fmt.Sprintf("Complaint CMP%d - Deduction Entry (No Payout)", complaint.InternalID),
		ComplaintID:         complaint.ID,
		ComplaintInternalID: complaint.InternalID,
	})

	if err != nil {
		fmt.Printf("ERROR: Failed to create deduction entry: %v", err)
		return []string{}
	}

	_ = s.acceptedServiceRepo.UpdateComplaintFlags(ctx, acceptedService.ID.Hex(), map[string]any{
		"payoutStatus":           domain.PayoutStatusComplaintAfterSettlement,
		"hasComplaintAdjustment": true,
		"pendingDeductionAmount": utils.RoundTo2(netPayable),
		"complaintId":            complaint.ID,
	})

	return []string{fmt.Sprintf("Deduction entry of %.2f created for future recovery", netPayable)}
}

func (s *ComplaintService) handleNoPayoutNotSettled(ctx context.Context, complaint *domain.Complaint, acceptedService *domain.AcceptedService, providerID string, originalAmount float64) []string {
	err := s.payoutService.ProcessPayout(ctx, dto.PayoutRequest{
		ProviderID:          providerID,
		BookingID:           complaint.AcceptedServiceID,
		Amount:              originalAmount,
		PartialAmount:       0,
		CancelPayout:        true,
		CreateDeduction:     false,
		Reason:              fmt.Sprintf("Complaint CMP%d - Payout Cancelled (No Payout)", complaint.InternalID),
		ComplaintID:         complaint.ID,
		ComplaintInternalID: complaint.InternalID,
	})

	if err != nil {
		fmt.Printf("ERROR: Failed to cancel payout: %v", err)
		return []string{}
	}

	_ = s.acceptedServiceRepo.UpdateComplaintFlags(ctx, acceptedService.ID.Hex(), map[string]any{
		"payoutStatus":      domain.PayoutStatusCancelled,
		"isPayoutCancelled": true,
		"PayoutCancelledAt": time.Now(),
	})

	return []string{"Existing payout cancelled - no payment to provider"}
}

func (s *ComplaintService) handlePartialPayout(ctx context.Context, complaint *domain.Complaint, acceptedService *domain.AcceptedService, providerID string, originalAmount, payoutAmount float64, isSettled bool) []string {
	err := s.payoutService.ProcessPayout(ctx, dto.PayoutRequest{
		ProviderID:          providerID,
		BookingID:           complaint.AcceptedServiceID,
		Amount:              originalAmount,
		PartialAmount:       payoutAmount,
		CancelPayout:        false,
		CreateDeduction:     false,
		Reason:              fmt.Sprintf("Complaint CMP%d", complaint.InternalID),
		ComplaintID:         complaint.ID,
		ComplaintInternalID: complaint.InternalID,
	})

	if err != nil {
		fmt.Printf("Warning: Failed to process payout: %v", err)
		return []string{}
	}

	if isSettled {
		deductionAmount := originalAmount - payoutAmount
		commission := deductionAmount * 20.0 / 100
		afterCommission := deductionAmount - commission
		gst := afterCommission * 18.0 / 100
		netDeduction := deductionAmount - commission - gst

		_ = s.acceptedServiceRepo.UpdateComplaintFlags(ctx, acceptedService.ID.Hex(), map[string]any{
			"payoutStatus":           domain.PayoutStatusComplaintAfterSettlement,
			"hasComplaintAdjustment": true,
			"pendingDeductionAmount": utils.RoundTo2(netDeduction),
			"complaintId":            complaint.ID,
		})
	}

	return []string{fmt.Sprintf("Payout of %.2f processed", payoutAmount)}
}

func (s *ComplaintService) StartAssessment(ctx context.Context, complaintID string) error {
	complaint, err := s.GetComplaint(ctx, complaintID)

	if err != nil {
		return fmt.Errorf("failed to get complaint: %w", err)
	}

	if complaint.Status != domain.ComplaintStatusInitiated {
		return fmt.Errorf("complaint must be in initiated status to start assessment, current status: %s", complaint.Status)
	}

	return s.complaintRepo.UpdateStatus(ctx, complaint.ID, domain.ComplaintStatusInReview)
}

func (s *ComplaintService) UpdateComplaintStatus(ctx context.Context, complaintID string, status domain.ComplaintStatus, adminID string) error {

	complaint, err := s.GetComplaint(ctx, complaintID)
	
	if err != nil {
		return fmt.Errorf("failed to get complaint: %w", err)
	}

	if !isValidComplaintStatus(status) {
		return fmt.Errorf("invalid complaint status: %s", status)
	}

	now := time.Now()
	updateData := bson.M{
		"updatedByAdmin": adminID,
		"adminUpdatedAt": now,
	}

	if err := s.complaintRepo.Update(ctx, complaint.ID, updateData); err != nil {
		return fmt.Errorf("failed to update admin info: %w", err)
	}

	if err := s.complaintRepo.UpdateStatus(ctx, complaint.ID, status); err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	return nil
}

func (s *ComplaintService) AddNote(ctx context.Context, internalID int64, req dto.AddNoteRequest) error {

	if req.Content == "" {
		return fmt.Errorf("note content cannot be empty")
	}

	if req.AddedBy == "" {
		return fmt.Errorf("addedBy field is required")
	}

	note := domain.ComplaintNote{
		ID:        primitive.NewObjectID().Hex(),
		Content:   req.Content,
		AddedBy:   req.AddedBy,
		CreatedAt: time.Now(),
	}

	return s.complaintRepo.AddNote(ctx, internalID, note)

}

func (s *ComplaintService) GetComplaintStats(ctx context.Context) (*dto.ComplaintStats, error) {
	return s.complaintRepo.GetStats(ctx)
}

func isValidComplaintStatus(status domain.ComplaintStatus) bool {
	validStatuses := []domain.ComplaintStatus{
		domain.ComplaintStatusInitiated,
		domain.ComplaintStatusInReview,
		domain.ComplaintStatusResolved,
		domain.ComplaintStatusCancelled,
	}

	for _, s := range validStatuses {
		if domain.ComplaintStatus(s) == status {
			return true
		}
	}
	return false
}
