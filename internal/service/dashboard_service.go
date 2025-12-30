package service

import (
	"context"
	"provider_management/internal/dto"
	"provider_management/internal/repository"
     "provider_management/internal/utils"
)

type DashboardService struct {
	providers    *repository.ProviderRepo
	users        *repository.UserRepo
	services     *repository.AcceptedServiceRepo
	settlements  *repository.ProviderSettlementRepo
	complaints   *repository.ComplaintRepository
	transactions *repository.TransactionRepo
	amcPurchases *repository.AMCPurchaseRepo
	bids         *repository.BidRepo
}

func NewDashboardService(
	providers *repository.ProviderRepo,
	users *repository.UserRepo,
	services *repository.AcceptedServiceRepo,
	settlements *repository.ProviderSettlementRepo,
	complaints *repository.ComplaintRepository,
	transactions *repository.TransactionRepo,
	amcPurchases *repository.AMCPurchaseRepo,
	bids         *repository.BidRepo,
) *DashboardService {
	return &DashboardService{
		providers:    providers,
		users:        users,
		services:     services,
		settlements:  settlements,
		complaints:   complaints,
		transactions: transactions,
		amcPurchases: amcPurchases,
		bids:         bids,
	}
}

func (s *DashboardService) GetDashboardStats(ctx context.Context, period string, days string) (*dto.DashboardResponse, error) {
    
	providers, err := s.providers.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	users, err := s.users.GetStats(ctx)
	if err != nil {
		return nil, err
	}

	amcCount, err := s.amcPurchases.CountActive(ctx)
	if err != nil {
		return nil, err
	}
	users.AMC = amcCount

	biddings, err := s.bids.GetBiddingStats(ctx)
	if err != nil {
		return nil, err
	}

	bookings, err := s.services.GetBookingStats(ctx)
	if err != nil {
		return nil, err
	}

	transactionStats, err := s.transactions.GetTransactionStats(ctx, days)
    if err != nil {
        return nil, err
    }

	settlementStats, err := s.settlements.GetStats(ctx, days)
    if err != nil {
        return nil, err
    }
	platformRevenue := utils.RoundTo2(transactionStats.TotalAmount - settlementStats.SettledAmount)
	baseAmount := utils.RoundTo2(transactionStats.TotalAmount - transactionStats.GSTAmount)
    
    var settledPercentage, platformPercentage, gstPercentage float64
    
    if baseAmount > 0 {
        settledPercentage = utils.RoundTo2((settlementStats.SettledAmount / baseAmount) * 100)
        platformPercentage = utils.RoundTo2((platformRevenue / baseAmount) * 100)
        gstPercentage = utils.RoundTo2((transactionStats.GSTAmount / baseAmount) * 100)
    } else {
        settledPercentage = 0
        platformPercentage = 0
        gstPercentage = 0
    }
	combinedSettlement := dto.CombinedSettlementStats{
        TransactionStats: transactionStats,
        SettlementStats:  settlementStats,
        PlatformRevenue:  platformRevenue,
		SettledPercentage: settledPercentage,
        PlatformPercentage: platformPercentage,
        GSTPercentage: gstPercentage,
    }
    

	complaints, err := s.complaints.GetDashboardStats(ctx)
	if err != nil {
		return nil, err
	}

	revenue, err := s.transactions.GetAMCRevenueStats(ctx, period)
	if err != nil {
		return nil, err
	}

	topServices, err := s.services.GetTopServices(ctx)
	if err != nil {
		return nil, err
	}

	return &dto.DashboardResponse{
		Providers:   providers,
		Users:       users,
		Biddings:    biddings,
		Bookings:    bookings,
		Settlement:  combinedSettlement, 
		Complaints:  complaints,
		Revenue:     revenue,
		TopServices: topServices,
	}, nil
}