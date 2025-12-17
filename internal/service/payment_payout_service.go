package service

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"strconv"
	"strings"
	"time"
)

type PayoutService struct {
	serviceRepo *repository.AcceptedServiceRepo
	payoutRepo  *repository.PaymentPayoutRepo
}

func NewPayoutService(serviceRepo *repository.AcceptedServiceRepo, payoutRepo *repository.PaymentPayoutRepo) *PayoutService {
	return &PayoutService{
		serviceRepo: serviceRepo,
		payoutRepo:  payoutRepo,
	}
}

func (s *PayoutService) CreatePayoutLast6Hours(ctx context.Context) error {

	to := time.Now()
	from := to.Add(-30 * 24 * time.Hour)

	services, err := s.serviceRepo.FindCompletedPaidBetween(ctx, from, to)
	if err != nil {
		return err
	}

	type bucket struct {
		ServiceIDs []primitive.ObjectID
		Total      float64
	}

	group := make(map[string]*bucket)

	for _, svc := range services {
		if _, ok := group[svc.ProviderID.Hex()]; !ok {
			group[svc.ProviderID.Hex()] = &bucket{}
		}

		objID, _ := primitive.ObjectIDFromHex(svc.ID.Hex())

		group[svc.ProviderID.Hex()].ServiceIDs = append(
			group[svc.ProviderID.Hex()].ServiceIDs,
			objID,
		)

		group[svc.ProviderID.Hex()].Total += svc.FinalPrice
	}

	for providerID, data := range group {
		payoutID := time.Now().UnixMilli()
		base := data.Total
		commissionPercent := 20.0
		gstPercent := 18.0

		commission := base * commissionPercent / 100
		gst := commission * gstPercent / 100
		net := base - commission - gst

		providerObjID, _ := primitive.ObjectIDFromHex(providerID)

		payout := &domain.PaymentPayout{
			PayoutID:          payoutID,
			ProviderID:        providerObjID,
			ServiceIDs:        data.ServiceIDs,
			BaseAmount:        base,
			CommissionPercent: commissionPercent,
			CommissionAmount:  commission,
			GSTPercent:        gstPercent,
			GSTAmount:         gst,
			NetPayable:        net,
			Status:            "pending",
			PeriodFrom:        from,
			PeriodTo:          to,
			CreatedAt:         time.Now(),
		}

		if err := s.payoutRepo.Create(ctx, payout); err != nil {
			return err
		}
	}

	return nil
}

func (s *PayoutService) GetPayouts(
	ctx context.Context,
	page, limit int64,
	search, providerID, status, periodFromStr, periodToStr, sortBy, sortOrder string,
) ([]map[string]any, int64, int64, error) {

	skip := (page - 1) * limit
	filter := bson.M{}

	if providerID != "" {
		if objID, err := primitive.ObjectIDFromHex(providerID); err == nil {
			filter["providerId"] = objID
		}
	}

	if status != "" {
		filter["status"] = status
	}

	if periodFromStr != "" && periodToStr != "" {
		from, err1 := time.Parse(time.RFC3339, periodFromStr)
		to, err2 := time.Parse(time.RFC3339, periodToStr)
		if err1 == nil && err2 == nil {
			filter["createdAt"] = bson.M{"$gte": from, "$lte": to}
		}
	}

	if search != "" {
		if objID, err := primitive.ObjectIDFromHex(search); err == nil {
			filter["providerId"] = objID
		} else {
			filter["$or"] = []bson.M{
				{"status": bson.M{"$regex": search, "$options": "i"}},
			}
		}
	}

	order := -1
	if sortOrder == "asc" {
		order = 1
	}

	if sortBy == "" {
		sortBy = "updatedAt"
	} else {
		switch sortBy {
		case "payout_id":
			sortBy = "payoutId"
		case "provider_id":
			sortBy = "providerId"
		case "base_amount":
			sortBy = "baseAmount"
		case "net_payable":
			sortBy = "netPayable"
		case "status":
			sortBy = "status"
		case "created_at":
			sortBy = "createdAt"
		case "updated_at":
			sortBy = "updatedAt"
		default:
			sortBy = sortBy
		}
	}

	payouts, total, err := s.payoutRepo.GetPayouts(ctx, filter, skip, limit, sortBy, order)
	if err != nil {
		return nil, 0, 0, err
	}

	responseData := make([]map[string]any, len(payouts))
	for i, p := range payouts {
		responseData[i] = map[string]any{
			"id":                 p.ID.Hex(),
			"payout_id":          "SET" + strconv.FormatInt(p.PayoutID, 10),
			"provider_id":        p.ProviderID.Hex(),
			"service_ids":        p.ServiceIDs,
			"base_amount":        p.BaseAmount,
			"commission_percent": p.CommissionPercent,
			"commission_amount":  p.CommissionAmount,
			"gst_percent":        p.GSTPercent,
			"gst_amount":         p.GSTAmount,
			"net_payable":        p.NetPayable,
			"status":             p.Status,
			"period_from":        p.PeriodFrom,
			"period_to":          p.PeriodTo,
			"created_at":         p.CreatedAt,
			"updated_at":         p.UpdatedAt,
		}
	}

	totalPages := total / limit
	if total%limit > 0 {
		totalPages++
	}

	return responseData, total, totalPages, nil
}

func (s *PayoutService) GetPayoutServices(ctx context.Context, payoutID string) ([]map[string]any, error) {
	numericID := payoutID
	if strings.HasPrefix(strings.ToUpper(payoutID), "SET") {
		numericID = payoutID[3:]
	}

	payoutIDInt, err := strconv.ParseInt(numericID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid payout ID: %v", err)
	}

	filter := bson.M{"payoutId": payoutIDInt}

	payouts, _, err := s.payoutRepo.GetPayouts(ctx, filter, 0, 1, "createdAt", -1)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payout: %v", err)
	}

	if len(payouts) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	payout := payouts[0]

	var services []map[string]any
	for _, serviceID := range payout.ServiceIDs {
		service, err := s.serviceRepo.FindByID(ctx, serviceID.Hex())
		if err != nil {
			log.Printf("Error fetching service %s: %v", serviceID.Hex(), err)
			continue
		}

		serviceCommission := service.FinalPrice * payout.CommissionPercent / 100
		serviceGST := serviceCommission * payout.GSTPercent / 100
		serviceNet := service.FinalPrice - serviceCommission - serviceGST

		serviceData := map[string]any{
			"id": service.ID,
			"booking_id":         fmt.Sprintf("BK%d", service.InternalID),
			"amc_id":             "amc",
			"provider_id":        payout.ProviderID.Hex(),
			"service_amount":     service.FinalPrice,
			"commission_percent": payout.CommissionPercent,
			"commission_amount":  serviceCommission,
			"gst_percent":        payout.GSTPercent,
			"gst_amount":         serviceGST,
			"net_amount":         serviceNet,
			"partial_amount":     "₹",
			"payout_id":          fmt.Sprintf("SET%d", payout.PayoutID),
			"is_settled":     service.IsSettled,
			"settlement_id":  service.SettlementID,
			"settled_at":     service.SettledAt,
		}

		services = append(services, serviceData)
	}

	return services, nil
}

func (s *PayoutService) GetPayoutProviderData(ctx context.Context, payoutID string) ([]map[string]any, error) {
	numericID := payoutID
	if strings.HasPrefix(strings.ToUpper(payoutID), "SET") {
		numericID = payoutID[3:]
	}

	payoutIDInt, err := strconv.ParseInt(numericID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid payout ID: %v", err)
	}

	filter := bson.M{"payoutId": payoutIDInt}

	payouts, _, err := s.payoutRepo.GetPayouts(ctx, filter, 0, 1, "createdAt", -1)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch payout: %v", err)
	}
    log.Println("payoutDataaa",payouts);
	if len(payouts) == 0 {
		return nil, mongo.ErrNoDocuments
	}

	payout := payouts[0]

	var services []map[string]any
	for _, serviceID := range payout.ServiceIDs {
		service, err := s.serviceRepo.FindByID(ctx, serviceID.Hex())
		if err != nil {
			log.Printf("Error fetching service %s: %v", serviceID.Hex(), err)
			continue
		}

		
		serviceData := map[string]any{
			"booking_id":         fmt.Sprintf("BK%d", service.InternalID),
			"amc_id":             "amc",
			"provider_id":        payout.ProviderID.Hex(),
			"service_amount":     service.FinalPrice,
			"commission_percent": payout.CommissionPercent,
			"gst_percent":        payout.GSTPercent,
			"partial_amount":     "₹",
			"payout_id":          fmt.Sprintf("SET%d", payout.PayoutID),
		}

		services = append(services, serviceData)
	}

	return services, nil
}
