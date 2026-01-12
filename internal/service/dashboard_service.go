package service

import (
	"context"
	"sync"

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

func (s *DashboardService) GetDashboardStats(ctx context.Context,period string,days string) (*dto.DashboardResponse, error) {

	var (
		providers        dto.ProvidersStats
		users            dto.UsersStats
		biddings         dto.BiddingStats
		bookings         dto.BookingsStats
		transactionStats dto.TransactionStats
		settlementStats  dto.SettlementStats
		complaints       dto.ComplaintsStats
		revenue          dto.RevenueStats
		topServices      []dto.TopService
		amcCount         int64
	)

	var (
		wg  sync.WaitGroup
		err error
		mu  sync.Mutex
	)

	setErr := func(e error) {
		if e != nil {
			mu.Lock()
			if err == nil {
				err = e
			}
			mu.Unlock()
		}
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		var e error
		providers, e = s.providers.GetStats(ctx)
		setErr(e)
	}()

	// Users
	wg.Add(1)
	go func() {
		defer wg.Done()
		var e error
		users, e = s.users.GetStats(ctx)
		setErr(e)
	}()

	// AMC
	wg.Add(1)
	go func() {
		defer wg.Done()
		var e error
		amcCount, e = s.amcPurchases.CountActive(ctx)
		setErr(e)
	}()

	// Biddings
	wg.Add(1)
	go func() {
		defer wg.Done()
		var e error
		biddings, e = s.bids.GetBiddingStats(ctx)
		setErr(e)
	}()

	// Bookings
	wg.Add(1)
	go func() {
		defer wg.Done()
		var e error
		bookings, e = s.services.GetBookingStats(ctx)
		setErr(e)
	}()

	// Transactions
	wg.Add(1)
	go func() {
		defer wg.Done()
		var e error
		transactionStats, e = s.transactions.GetTransactionStats(ctx, days)
		setErr(e)
	}()

	// Settlements
	wg.Add(1)
	go func() {
		defer wg.Done()
		var e error
		settlementStats, e = s.settlements.GetStats(ctx, days)
		setErr(e)
	}()

	// Complaints
	wg.Add(1)
	go func() {
		defer wg.Done()
		var e error
		complaints, e = s.complaints.GetDashboardStats(ctx)
		setErr(e)
	}()

	// Revenue
	wg.Add(1)
	go func() {
		defer wg.Done()
		var e error
		revenue, e = s.transactions.GetAMCRevenueStats(ctx, period)
		setErr(e)
	}()

	// Top Services
	wg.Add(1)
	go func() {
		defer wg.Done()
		var e error
		topServices, e = s.services.GetTopServices(ctx)
		setErr(e)
	}()

	wg.Wait()
	if err != nil {
		return nil, err
	}

	users.AMC = amcCount

	platformRevenue := utils.RoundTo2(
		transactionStats.TotalAmount - settlementStats.SettledAmount,
	)

	baseAmount := utils.RoundTo2(
		transactionStats.TotalAmount - transactionStats.GSTAmount,
	)

	var settledPercentage, platformPercentage, gstPercentage float64
	if baseAmount > 0 {
		settledPercentage = utils.RoundTo2((settlementStats.SettledAmount / baseAmount) * 100)
		platformPercentage = utils.RoundTo2((platformRevenue / baseAmount) * 100)
		gstPercentage = utils.RoundTo2((transactionStats.GSTAmount / baseAmount) * 100)
	}

	combinedSettlement := dto.CombinedSettlementStats{
		TransactionStats:   transactionStats,
		SettlementStats:    settlementStats,
		PlatformRevenue:    platformRevenue,
		SettledPercentage:  settledPercentage,
		PlatformPercentage: platformPercentage,
		GSTPercentage:      gstPercentage,
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
