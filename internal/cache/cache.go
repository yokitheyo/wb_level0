package cache

import (
	"sync"

	"github.com/yokitheyo/wb_level0/internal/models"
	"go.uber.org/zap"
)

type OrderCache interface {
	Set(orderUID string, order models.Order)
	Get(orderUID string) (models.Order, bool)
	LoadFromDB(orders []models.Order)
	GetStats() CacheStats
}

type CacheStats struct {
	TotalOrders int      `json:"total_orders"`
	OrderUIDs   []string `json:"order_uids"`
}

type orderCache struct {
	mu     sync.RWMutex
	cache  map[string]models.Order
	logger *zap.Logger
}

func NewOrderCache(logger *zap.Logger) OrderCache {
	return &orderCache{
		cache:  make(map[string]models.Order),
		logger: logger,
	}
}

func (c *orderCache) Set(orderUID string, order models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[orderUID] = order
	c.logger.Debug("order added to cache", zap.String("order_uid", orderUID))
}

func (c *orderCache) Get(orderUID string) (models.Order, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	order, ok := c.cache[orderUID]
	return order, ok
}

func (c *orderCache) LoadFromDB(orders []models.Order) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, order := range orders {
		c.cache[order.OrderUID] = order
	}

	c.logger.Info("cache loaded from database", zap.Int("orders_count", len(orders)))
}

func (c *orderCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	orderUIDs := make([]string, 0, len(c.cache))
	for uid := range c.cache {
		orderUIDs = append(orderUIDs, uid)
	}

	return CacheStats{
		TotalOrders: len(c.cache),
		OrderUIDs:   orderUIDs,
	}
}
