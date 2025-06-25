package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yokitheyo/wb_level0/internal/cache"
	"github.com/yokitheyo/wb_level0/internal/models"
	"github.com/yokitheyo/wb_level0/internal/repository"
	"go.uber.org/zap"
)

type OrderService interface {
	ProcessOrder(ctx context.Context, data []byte) error
	GetOrderByID(ctx context.Context, orderUID string) (*models.Order, error)
	RestoreCache(ctx context.Context) error
	GetCacheStats() cache.CacheStats
}

type orderService struct {
	repo   repository.OrderRepository
	cache  cache.OrderCache
	logger *zap.Logger
}

func NewOrderService(repo repository.OrderRepository,
	cache cache.OrderCache,
	logger *zap.Logger) OrderService {
	return &orderService{
		repo:   repo,
		cache:  cache,
		logger: logger,
	}
}

func (s *orderService) ProcessOrder(ctx context.Context, data []byte) error {
	var order models.Order
	if err := json.Unmarshal(data, &order); err != nil {
		s.logger.Error("failed to unmarshal order", zap.Error(err), zap.String("data", string(data)))
		return fmt.Errorf("failed to unmarshal order: %w", err)
	}

	if err := s.validateOrder(order); err != nil {
		s.logger.Error("invalid order data", zap.Error(err), zap.String("order_uid", order.OrderUID))
		return fmt.Errorf("invalid order data: %w", err)
	}

	if err := s.repo.SaveOrder(ctx, order); err != nil {
		s.logger.Error("failed to save order", zap.Error(err), zap.String("order_uid", order.OrderUID))
		return fmt.Errorf("failed to save order: %w", err)
	}

	s.cache.Set(order.OrderUID, order)
	s.logger.Info("order processed successfully", zap.String("order_uid", order.OrderUID))
	return nil
}

func (s *orderService) GetOrderByID(ctx context.Context, orderUID string) (*models.Order, error) {
	if order, ok := s.cache.Get(orderUID); ok {
		s.logger.Debug("order found in cache", zap.String("order_uid", orderUID))
		return &order, nil
	}

	s.logger.Debug("order not found in cache, trying database", zap.String("order_uid", orderUID))
	order, err := s.repo.GetOrderByID(ctx, orderUID)
	if err != nil {
		s.logger.Error("failed to get order from database", zap.Error(err), zap.String("order_uid", orderUID))
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	if order != nil {
		s.cache.Set(orderUID, *order)
	}
	return order, nil
}

func (s *orderService) RestoreCache(ctx context.Context) error {
	s.logger.Info("restoring cache from database")
	orders, err := s.repo.GetAllOrders(ctx)
	if err != nil {
		s.logger.Error("failed to get orders from database", zap.Error(err))
		return fmt.Errorf("failed to get orders: %w", err)
	}
	s.cache.LoadFromDB(orders)
	s.logger.Info("orders restored successfully", zap.Int("num_orders", len(orders)))
	return nil
}

func (s *orderService) GetCacheStats() cache.CacheStats {
	return s.cache.GetStats()
}

func (s *orderService) validateOrder(order models.Order) error {
	if order.OrderUID == "" {
		return fmt.Errorf("missing order_uid")
	}

	if order.TrackNumber == "" {
		return fmt.Errorf("missing track_number")
	}

	if order.Entry == "" {
		return fmt.Errorf("missing entry")
	}

	if order.CustomerID == "" {
		return fmt.Errorf("missing customer_id")
	}

	if order.DeliveryService == "" {
		return fmt.Errorf("missing delivery_service")
	}

	if err := s.validateDelivery(order.Delivery); err != nil {
		return fmt.Errorf("invalid delivery: %w", err)
	}

	if err := s.validatePayment(order.Payment); err != nil {
		return fmt.Errorf("invalid payment: %w", err)
	}

	if len(order.Items) == 0 {
		return fmt.Errorf("no items in order")
	}

	for i, item := range order.Items {
		if err := s.validateItem(item); err != nil {
			return fmt.Errorf("invalid item at index %d: %w", i, err)
		}
	}

	return nil
}

func (s *orderService) validateDelivery(delivery models.Delivery) error {
	if delivery.Name == "" {
		return fmt.Errorf("missing delivery name")
	}

	if delivery.Phone == "" {
		return fmt.Errorf("missing delivery phone")
	}

	if delivery.City == "" {
		return fmt.Errorf("missing delivery city")
	}

	if delivery.Address == "" {
		return fmt.Errorf("missing delivery address")
	}

	if delivery.Email != "" && !strings.Contains(delivery.Email, "@") {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

func (s *orderService) validatePayment(payment models.Payment) error {
	if payment.Transaction == "" {
		return fmt.Errorf("missing payment transaction")
	}

	if payment.Currency == "" {
		return fmt.Errorf("missing payment currency")
	}

	if payment.Provider == "" {
		return fmt.Errorf("missing payment provider")
	}

	if payment.Amount <= 0 {
		return fmt.Errorf("invalid payment amount: %d", payment.Amount)
	}

	if payment.Bank == "" {
		return fmt.Errorf("missing payment bank")
	}

	return nil
}

func (s *orderService) validateItem(item models.Item) error {
	if item.ChrtID <= 0 {
		return fmt.Errorf("invalid chrt_id: %d", item.ChrtID)
	}

	if item.TrackNumber == "" {
		return fmt.Errorf("missing item track_number")
	}

	if item.Name == "" {
		return fmt.Errorf("missing item name")
	}

	if item.Price <= 0 {
		return fmt.Errorf("invalid item price: %d", item.Price)
	}

	if item.TotalPrice <= 0 {
		return fmt.Errorf("invalid item total_price: %d", item.TotalPrice)
	}

	if item.Brand == "" {
		return fmt.Errorf("missing item brand")
	}

	return nil
}
