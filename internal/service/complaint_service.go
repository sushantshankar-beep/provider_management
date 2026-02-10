package service

import (
	"context"
	"fmt"
	"log"
	"provider_management/internal/domain"
	"provider_management/internal/dto"
	"provider_management/internal/repository"
	"provider_management/internal/utils"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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
	transactionRepo     *repository.TransactionRepo
	kycRepo             *repository.ProviderKYCRepository
}

func NewComplaintService(
	paymentPayoutRepo *repository.PaymentPayoutRepo,
	complaintRepo *repository.ComplaintRepository,
	acceptedServiceRepo *repository.AcceptedServiceRepo,
	userRepo *repository.UserRepo,
	providerRepo *repository.ProviderRepo,
	refundService *RefundService,
	payoutService *PayoutService,
	transactionRepo     *repository.TransactionRepo,
	kycRepo             *repository.ProviderKYCRepository,
) *ComplaintService {
	return &ComplaintService{
		paymentPayoutRepo:   paymentPayoutRepo,
		complaintRepo:       complaintRepo,
		acceptedServiceRepo: acceptedServiceRepo,
		userRepo:            userRepo,
		providerRepo:        providerRepo,
		refundService:       refundService,
		payoutService:       payoutService,
		transactionRepo: transactionRepo,
		kycRepo:kycRepo,
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

func (s *ComplaintService) GetComplaintByNumber(ctx context.Context, complaintNumber string ) (*domain.Complaint, error) {
	return s.complaintRepo.GetByComplaintNumber(ctx, complaintNumber)
}



func (s *ComplaintService) GetAllComplaints(
	ctx context.Context,
	filters dto.ComplaintFilters,
	pagination dto.ComplaintPagination,
) (*dto.ComplaintListResponse, error) {

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
		var booking *domain.AcceptedService
		if complaint.AcceptedService != "" {
			booking, err = s.acceptedServiceRepo.FindByID(ctx, complaint.AcceptedService)
			if err != nil {
				log.Println(
					"Error fetching accepted service:",
					err,
					"AcceptedServiceID:",
					complaint.AcceptedService,
				)
			}
		}

		serviceNumber := ""
		if booking != nil {
			serviceNumber = booking.ServiceNumber
		}

		resp := dto.ComplaintListItem{
			ID:                complaint.ID,
			InternalID:        complaint.ComplaintNumber,
			AcceptedServiceNo: serviceNumber,
			Status:            complaint.Status,
			CreatedAt:         indianTime.Format("2006-01-02 15:04:05"),
		}

		if complaint.UserID != "" {
			user, err := s.userRepo.FindByID(ctx, complaint.UserID)
			if err != nil {
				log.Println("Error fetching user:", err, "UserID:", complaint.UserID)
			} else {
				resp.UserName = user.Name
			}
		}

		if complaint.ProviderID != "" {
			provider, err := s.providerRepo.FindByID(ctx, complaint.ProviderID)
			if err != nil {
				log.Println("Error fetching provider:", err, "ProviderID:", complaint.ProviderID)
			} else {
				resp.ProviderName = provider.Name
			}
		}

		if complaint.UserComplaint != nil {
			resp.RaisedBy = "user"
			resp.Against = "provider"
		}

		if complaint.ProviderComplaint != nil {
			resp.RaisedBy = "provider"
			resp.Against = "user"
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

func (s *ComplaintService) GetComplaintWithDetails(
	ctx context.Context,
	id string,
) (*dto.ComplaintDetailResponse, error) {

	complaint, err := s.complaintRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	var booking *domain.AcceptedService
	if complaint.AcceptedService != "" {
		booking, _ = s.acceptedServiceRepo.FindByID(
			ctx,
			complaint.AcceptedService,
		)
	}

	var user *domain.User
	if complaint.UserID != "" {
		log.Println("kdsnjksnjkvd",complaint.UserID)
		user, err = s.userRepo.FindByID(ctx, complaint.UserID)
		log.Println("dkjsncjksnac",user)
		if err != nil {
			log.Println("Error fetching user:", err, "UserID:", complaint.UserID)
		}
	}

	var provider *domain.Provider
	if complaint.ProviderID != "" {
		provider, err = s.providerRepo.FindByID(ctx, complaint.ProviderID)
		if err != nil {
			log.Println("Error fetching provider:", err, "ProviderID:", complaint.ProviderID)
		}
	}

	var transaction *domain.Transaction
	if complaint.AcceptedService != "" {
		transaction, err = s.transactionRepo.FindByServiceID(ctx, complaint.AcceptedService)
		if err != nil {
			log.Println("Error fetching transactio:", err, "Service", complaint.AcceptedService)
		}
	}

	resp := &dto.ComplaintDetailResponse{
		ComplaintID:     complaint.ID,
		ComplaintNumber: complaint.ComplaintNumber,
		Status:          complaint.Status,
		Tracking:        complaint.Timeline,
		Notes:  complaint.Notes,
		Assessment:       complaint.Assessment,
		ActionsTriggered: complaint.ActionsTriggered,
		CreatedAt:        complaint.CreatedAt,
		UpdatedAt:        complaint.UpdatedAt,
		UpdatedByAdmin:   complaint.UpdatedByAdmin,
		AdminUpdatedAt:   complaint.AdminUpdatedAt,
		PaymentTracking:  complaint.PaymentTracking,
	}

	resp.ComplaintInformation = dto.ComplaintInformation{
		ComplaintID: complaint.ComplaintNumber,
		SubmittedAt: complaint.CreatedAt,
	}

	if booking != nil {
		resp.ComplaintInformation.BookingID = booking.ServiceNumber
		resp.ComplaintInformation.BookingAmount = transaction.Amount
	}

	if complaint.UserComplaint != nil {
		userInfo := dto.PartyInfo{Role: "User"}
		if user != nil {
			userInfo.ID = user.ID
			userInfo.InternalID = user.UserCode
			userInfo.Name = user.Name
			userInfo.Mobile = user.Phone
		}

		providerInfo := dto.PartyInfo{Role: "Provider"}
		if provider != nil {
			providerInfo.ID = provider.ID.Hex()
			providerInfo.InternalID = provider.ProviderCode
			providerInfo.Name = provider.Name
			providerInfo.Mobile = provider.Phone
		}

		resp.UserToProvider = &dto.ComplaintSideUI{
			SubmittedBy:      userInfo,
			ComplaintAgainst: providerInfo,
			Description:      complaint.UserComplaint.Problem,
			Images:           complaint.UserComplaint.Photos,
			RaisedAt:         complaint.UserComplaint.RaisedAt,
		}
	}

	if complaint.ProviderComplaint != nil {
		providerInfo := dto.PartyInfo{Role: "Provider"}
		if provider != nil {
			providerInfo.ID = provider.ID.Hex()
			providerInfo.InternalID = provider.ProviderCode
			providerInfo.Name = provider.Name
			providerInfo.Mobile = provider.Phone
		}

		userInfo := dto.PartyInfo{Role: "User"}
		if user != nil {
			userInfo.ID = user.ID
			userInfo.InternalID = user.UserCode
			userInfo.Name = user.Name
			userInfo.Mobile = user.Phone
		}

		resp.ProviderToUser = &dto.ComplaintSideUI{
			SubmittedBy:      providerInfo,
			ComplaintAgainst: userInfo,
			Description:      complaint.ProviderComplaint.Problem,
			Images:           complaint.ProviderComplaint.Photos,
			RaisedAt:         complaint.ProviderComplaint.RaisedAt,
		}
	}

	return resp, nil
}

func (s *ComplaintService) AssessComplaint(ctx context.Context, complaintID string, req dto.AssessComplaintRequest) error {
	complaint, err := s.GetComplaintByNumber(ctx, complaintID)

	if err != nil {
		return fmt.Errorf("failed to get complaint: %w", err)
	}

	if complaint.Status != domain.ComplaintStatusInReview {
		return fmt.Errorf("complaint must be in in_review status to save assessment, current status: %s", complaint.Status)
	}

	acceptedService, err := s.acceptedServiceRepo.FindByID(ctx, complaint.AcceptedService)

	if err != nil {
		return fmt.Errorf("failed to get accepted service: %w", err)
	}

	if acceptedService.PaymentStatus != "paid" {
		return fmt.Errorf("no transaction associated with this booking")
	}

	// var transaction *domain.Transaction
	// if complaint.AcceptedService != "" {
	// 	transaction, err = s.transactionRepo.FindByServiceID(ctx, complaint.AcceptedService)
	// 	if err != nil {
	// 		log.Println("Error fetching transactio:", err, "Service", complaint.AcceptedService)
	// 	}
	// }

	if err := s.validateComplaintAssessmentAmounts(&req, acceptedService.FinalPrice); err != nil {
		return err
	}	

	if err := s.saveAssessmentAndResolve(ctx, complaint.ID, req); err != nil {
		return err
	}

	return s.processPaymentActions(ctx, complaint, acceptedService, req)

}

func (s *ComplaintService) validateComplaintAssessmentAmounts(req *dto.AssessComplaintRequest, originalAmount float64) error {

	if originalAmount <= 0 {
		return fmt.Errorf("invalid original booking amount: %.2f", originalAmount)
	}

	log.Println("kcencjknwjedcw",originalAmount)

	// gstAmount, tdsAmount, netAmount := s.calculateNetAmounts(originalAmount)

	if req.RefundToUser == domain.RefundTypeFull {
		req.RefundAmount = originalAmount
	}
    log.Println("dkjsnjkcsndsc",req.RefundAmount)
	if req.PayoutToProvider == domain.PayoutTypeFull {
		req.PayoutAmount = originalAmount
	}

	log.Println("dkjsnjkcsndsc",req.PayoutAmount)

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
    log.Println("hcsdbhjsdcc",req.RefundAmount)
	log.Println("ndcsjkncds",req.PayoutAmount)
	assessment := domain.ComplaintAssessment{
		FaultParty:       req.FaultParty,
		RefundToUser:     req.RefundToUser,
		RefundAmount:     req.RefundAmount,
		PayoutToProvider: req.PayoutToProvider,
		PayoutAmount:     req.PayoutAmount,
		RemarkForUser:    req.RemarkForUser,
		RemarkForProvider:req.RemarkForProvider,
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
		refundAction := s.processRefund(ctx, complaint, acceptedService, req)
		if refundAction != "" {
			paymentTracking.RefundStatus = domain.PaymentActionPending
			actions = append(actions, refundAction)
		}
	}

	if !acceptedService.Provider.IsZero() {
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

func (s *ComplaintService) processRefund(ctx context.Context, complaint *domain.Complaint, acceptedService *domain.AcceptedService, req dto.AssessComplaintRequest) string {
    
	userObjID, err := primitive.ObjectIDFromHex(complaint.UserID)

	if err != nil {
		fmt.Printf("Warning: Invalid UserID format: %s", complaint.UserID)
		return ""
	}

	refundReason := req.RemarkForUser

	if req.RefundToUser == domain.RefundTypeFull {
		refundReason = refundReason
	} else {
		refundReason = refundReason
	}


	var transaction *domain.Transaction
	if complaint.AcceptedService != "" {
		transaction, err = s.transactionRepo.FindByServiceID(ctx, complaint.AcceptedService)
		if err != nil {
			log.Println("Error fetching transactio:", err, "Service", complaint.AcceptedService)
		}
	}

	complaintObjID, err := primitive.ObjectIDFromHex(complaint.ID)
   if err != nil {
     	log.Println("Invalid complaint ID:", complaint.ID)
	    return ""
    }

	if err := s.refundService.ProcessRefund(ctx, dto.RefundRequest{
		UserID:              userObjID.Hex(),
		TxnID:               transaction.TxnID,
		BookingID:           complaint.AcceptedService,
		ComplaintID:         complaintObjID.Hex(),
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

	// var transaction *domain.Transaction
	// if complaint.AcceptedService != "" {
	// 	var err error
	// 	transaction, err = s.transactionRepo.FindByServiceID(ctx, complaint.AcceptedService)
	// 	if err != nil {
	// 		log.Println("Error fetching transactio:", err, "Service", complaint.AcceptedService)
	// 	}
	// }

	providerID := acceptedService.Provider.Hex()
	originalAmount := acceptedService.FinalPrice
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
		BookingID:           complaint.AcceptedService,
		Amount:              originalAmount,
		PartialAmount:       0,
		CancelPayout:        false,
		CreateDeduction:     true,
		Reason:              fmt.Sprintf("Complaint CMP%d - Deduction Entry (No Payout)", complaint.ComplaintNumber),
		ComplaintID:         complaint.ID,
		ComplaintInternalID: complaint.ComplaintNumber,
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
		BookingID:           complaint.AcceptedService,
		Amount:              originalAmount,
		PartialAmount:       0,
		CancelPayout:        true,
		CreateDeduction:     false,
		Reason:              fmt.Sprintf("Complaint CMP%d - Payout Cancelled (No Payout)", complaint.ComplaintNumber),
		ComplaintID:         complaint.ID,
		ComplaintInternalID: complaint.ComplaintNumber,
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
	log.Println("jcsdbjbdjsbdcscds",originalAmount)
	log.Println("kjdsnckjnjksdcdcs",payoutAmount)
	err := s.payoutService.ProcessPayout(ctx, dto.PayoutRequest{
		ProviderID:          providerID,
		BookingID:           complaint.AcceptedService,
		Amount:              originalAmount,
		PartialAmount:       payoutAmount,
		CancelPayout:        false,
		CreateDeduction:     false,
		Reason:              fmt.Sprintf("Complaint CMP%d", complaint.ComplaintNumber),
		ComplaintID:         complaint.ID,
		ComplaintInternalID: complaint.ComplaintNumber,
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
	
	complaint, err := s.complaintRepo.GetByComplaintNumber(ctx, complaintID)

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

func (s *ComplaintService) AddNote(ctx context.Context, complaintID string, req dto.AddNoteRequest) error {

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

	return s.complaintRepo.AddNote(ctx, complaintID, note)

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

func (s *ComplaintService) calculateNetAmounts(originalAmount float64, hasGSTNumber bool) (gstAmount, tdsAmount, netAmount float64) {

    if hasGSTNumber {
        tdsAmount = originalAmount * 0.10
        netAmount = originalAmount - tdsAmount
        gstAmount = 0
    } else {
        gstAmount = originalAmount * 0.18
        netAmount = originalAmount - gstAmount
        tdsAmount = 0
    }
    return gstAmount, tdsAmount, netAmount
}

