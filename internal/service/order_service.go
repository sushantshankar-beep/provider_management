package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"provider_management/internal/domain"
	"provider_management/internal/repository"
	"strconv"
	"time"
)

type OrderService struct {
	orders   *repository.OrderRepo
	users    *repository.UserRepo
	plans    *repository.AMCPlanRepo
	vehicles *repository.SavedVehiclesRepo
	cities   *repository.ZoneRepo
}

func NewOrderService(
	o *repository.OrderRepo,
	u *repository.UserRepo,
	p *repository.AMCPlanRepo,
	v *repository.SavedVehiclesRepo,
	c *repository.ZoneRepo,
) *OrderService {
	return &OrderService{
		orders:   o,
		users:    u,
		plans:    p,
		vehicles: v,
		cities:   c,
	}
}

type OrderListResponse struct {
	Date       string  `json:"date"`
	FullName   string  `json:"fullName"`
	Contact    string  `json:"contact"`
	Email      string  `json:"email"`
	OrderID    string  `json:"orderId"`
	Amount     float64 `json:"amount"`
	Status     string  `json:"status"`
	PlanStatus string  `json:"planStatus"`
	ID         string  `json:"_id"`
}

type OrderDetailResponse struct {
	OrderDetails struct {
		Date     time.Time `json:"date"`
		FullName string    `json:"fullName"`
		Contact  string    `json:"contact"`
		Email    string    `json:"email"`
		OrderID  string    `json:"orderId"`
		Amount   float64   `json:"amount"`
		Message  string    `json:"message"`
	} `json:"orderDetails"`
	PlanDetails struct {
		PlanName           string  `json:"planName"`
		VehicleType        string  `json:"vehicleType"`
		PlanPrice          float64 `json:"planPrice"`
		GST                float64 `json:"gst"`
		PlanDiscountAmount float64 `json:"planDiscountAmount"`
		TotalPrice         float64 `json:"totalPrice"`
	} `json:"planDetails"`
	ServicesIncluded []domain.ServiceDetail `json:"servicesIncluded"`
	Vehicle          domain.VehicleInfo     `json:"vehicle"`
	PaymentDetails   struct {
		PayuID            int64  `json:"payuId"`
		PaymentID         string `json:"paymentId"`
		PayuTransactionID string `json:"payuTransactionId"`
		PaymentStatus     string `json:"paymentStatus"`
		PaymentSource     string `json:"paymentSource"`
	} `json:"paymentDetails"`
}

func (s *OrderService) ListOrders(
	ctx context.Context,
	pageStr, limitStr, search, status, paymentStatus, startDate, endDate, sortBy, sortOrder string,
) ([]OrderListResponse, int64, error) {

	page, _ := strconv.ParseInt(pageStr, 10, 64)
	limit, _ := strconv.ParseInt(limitStr, 10, 64)

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	skip := (page - 1) * limit

	orders, total, err := s.orders.FindWithFilter(
		ctx, skip, limit, search, status, paymentStatus,
		startDate, endDate, sortBy, sortOrder,
	)
	if err != nil {
		return nil, 0, err
	}

	var result []OrderListResponse

	for _, order := range orders {
		resp := OrderListResponse{
			Date:       order.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			OrderID:    order.ID.Hex(),
			Amount:     order.PlanPrice,
			Status:     order.PaymentStatus,
			PlanStatus: order.PlanStatus,
			ID:         order.ID.Hex(),
		}

		if order.UserID != primitive.NilObjectID {
			if user, err := s.users.FindByID(ctx, order.UserID.Hex()); err == nil {
				resp.FullName = user.Name
				resp.Contact = user.Phone
				resp.Email = user.Email
			} else {
				resp.FullName = "N/A"
				resp.Contact = "N/A"
				resp.Email = "N/A"
			}
		} else {
			resp.FullName = "N/A"
			resp.Contact = "N/A"
			resp.Email = "N/A"
		}

		result = append(result, resp)
	}

	return result, total, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*OrderDetailResponse, error) {
	order, err := s.orders.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := &OrderDetailResponse{}

	resp.OrderDetails.Date = order.CreatedAt
	resp.OrderDetails.OrderID = order.ID.Hex()
	resp.OrderDetails.Amount = order.PlanPrice
	if order.PaymentStatus == "success" {
		resp.OrderDetails.Message = "Transaction Completed Successfully"
	} else {
		resp.OrderDetails.Message = "Transaction " + order.PaymentStatus
	}

	if order.UserID != primitive.NilObjectID {
		if user, err := s.users.FindByID(ctx, order.UserID.Hex()); err == nil {
			resp.OrderDetails.FullName = user.Name
			resp.OrderDetails.Contact = user.Phone
			resp.OrderDetails.Email = user.Email
		} else {
			resp.OrderDetails.FullName = "N/A"
			resp.OrderDetails.Contact = "N/A"
			resp.OrderDetails.Email = "N/A"
		}
	}

	resp.PlanDetails.PlanName = order.PlanName
	resp.PlanDetails.VehicleType = order.Vehicle.VehicleType
	
	if order.PlanID != primitive.NilObjectID {
		if plan, err := s.plans.FindByID(ctx, order.PlanID.Hex()); err == nil {
			if order.PlanName == "" {
				resp.PlanDetails.PlanName = plan.PlanName
			}
			resp.PlanDetails.VehicleType = plan.PlanVehicleType
			resp.PlanDetails.PlanPrice = plan.PlanBasePrice
			resp.PlanDetails.GST = plan.PlanGSTAmount
			resp.PlanDetails.PlanDiscountAmount = plan.PlanDiscountAmount
			resp.PlanDetails.TotalPrice = plan.PlanTotalAmount
		}
	}

	resp.Vehicle = order.Vehicle

	resp.ServicesIncluded = order.ServiceDetails

	resp.PaymentDetails.PayuID = order.InternalID
	resp.PaymentDetails.PaymentID = order.PayuTransactionID
	resp.PaymentDetails.PayuTransactionID = order.PaymentID
	resp.PaymentDetails.PaymentStatus = order.PaymentStatus
	resp.PaymentDetails.PaymentSource = order.PaymentSource

	return resp, nil
}

func (s *OrderService) ExportOrdersToCSV(
	ctx context.Context,
	search, status, paymentStatus, startDate, endDate string,
) (string, error) {

	orders, _, err := s.orders.FindWithFilter(
		ctx, 0, 0, search, status, paymentStatus,
		startDate, endDate, "createdAt", "desc",
	)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	writer := csv.NewWriter(&buf)

	header := []string{
		"Date", "Full Name", "Contact", "Email", "Order ID",
		"Amount", "Payment Status", "Plan Status", "Plan Name",
		"Vehicle Number", "Vehicle Type",
	}
	if err := writer.Write(header); err != nil {
		return "", err
	}

	for _, order := range orders {
		var userName, userContact, userEmail string = "N/A", "N/A", "N/A"
		var planName string = "N/A"
		var vehicleNumber, vehicleType string = "N/A", "N/A"

		if order.UserID != primitive.NilObjectID {
			if user, err := s.users.FindByID(ctx, order.UserID.Hex()); err == nil {
				userName = user.Name
				userContact = user.Phone
				userEmail = user.Email
			}
		}

	
		planName = order.PlanName
		if planName == "" && order.PlanID != primitive.NilObjectID {
			if plan, err := s.plans.FindByID(ctx, order.PlanID.Hex()); err == nil {
				planName = plan.PlanName
			}
		}

		vehicleNumber = order.Vehicle.VehicleNumber
		vehicleType = order.Vehicle.VehicleType

		orderIDStr := order.PayuTransactionID
		if orderIDStr == "" {
			orderIDStr = order.ID.Hex()
		}

		row := []string{
			order.CreatedAt.Format("01/02/2006"),
			userName,
			userContact,
			userEmail,
			orderIDStr,
			fmt.Sprintf("%.2f", order.PlanPrice),
			order.PaymentStatus,
			order.PlanStatus,
			planName,
			vehicleNumber,
			vehicleType,
		}

		if err := writer.Write(row); err != nil {
			return "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (s *OrderService) UpdateOrderStatus(
	ctx context.Context,
	id, planStatus, paymentStatus string,
) (*domain.AMCPurchase, error) {
	return s.orders.UpdateStatus(ctx, id, planStatus, paymentStatus)
}