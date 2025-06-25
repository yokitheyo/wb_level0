package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/yokitheyo/wb_level0/internal/models"
	"go.uber.org/zap"
)

type OrderRepository interface {
	SaveOrder(ctx context.Context, order models.Order) error
	GetOrderByID(ctx context.Context, orderUID string) (*models.Order, error)
	GetAllOrders(ctx context.Context) ([]models.Order, error)
}

type orderRepository struct {
	db     *pgxpool.Pool
	logger *zap.Logger
}

func NewOrderRepository(db *pgxpool.Pool, logger *zap.Logger) OrderRepository {
	return &orderRepository{
		db:     db,
		logger: logger,
	}
}

func (r *orderRepository) SaveOrder(ctx context.Context, order models.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	_, err = tx.Exec(ctx, `
        INSERT INTO orders (
            order_uid, track_number, entry, locale, internal_signature,
            customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (order_uid) DO NOTHING`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.ShardKey, order.SmID, order.DateCreated, order.OofShard,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO deliveries (
            order_uid, name, phone, zip, city, address, region, email
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (order_uid) DO UPDATE SET
            name = EXCLUDED.name,
            phone = EXCLUDED.phone,
            zip = EXCLUDED.zip,
            city = EXCLUDED.city,
            address = EXCLUDED.address,
            region = EXCLUDED.region,
            email = EXCLUDED.email`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email,
	)
	if err != nil {
		return fmt.Errorf("failed to insert delivery: %w", err)
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO payments (
            order_uid, transaction, request_id, currency, provider,
            amount, payment_dt, bank, delivery_cost, goods_total, custom_fee
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        ON CONFLICT (order_uid) DO UPDATE SET
            transaction = EXCLUDED.transaction,
            request_id = EXCLUDED.request_id,
            currency = EXCLUDED.currency,
            provider = EXCLUDED.provider,
            amount = EXCLUDED.amount,
            payment_dt = EXCLUDED.payment_dt,
            bank = EXCLUDED.bank,
            delivery_cost = EXCLUDED.delivery_cost,
            goods_total = EXCLUDED.goods_total,
            custom_fee = EXCLUDED.custom_fee`,
		order.OrderUID, order.Payment.Transaction, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee,
	)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	for _, item := range order.Items {
		_, err = tx.Exec(ctx, `
            INSERT INTO items (
                order_uid, chrt_id, track_number, price, rid, name,
                sale, size, total_price, nm_id, brand, status
            ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.RID, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status,
		)
		if err != nil {
			return fmt.Errorf("failed to insert item: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *orderRepository) GetOrderByID(ctx context.Context, orderUID string) (*models.Order, error) {
	var order models.Order
	var delivery models.Delivery
	var payment models.Payment

	err := r.db.QueryRow(ctx, `
        SELECT 
            o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
            o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard
        FROM orders o
        WHERE o.order_uid = $1`,
		orderUID,
	).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
		&order.CustomerID, &order.DeliveryService, &order.ShardKey, &order.SmID, &order.DateCreated, &order.OofShard,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	err = r.db.QueryRow(ctx, `
        SELECT 
            name, phone, zip, city, address, region, email
        FROM deliveries
        WHERE order_uid = $1`,
		orderUID,
	).Scan(
		&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City,
		&delivery.Address, &delivery.Region, &delivery.Email,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get delivery: %w", err)
	}
	order.Delivery = delivery

	err = r.db.QueryRow(ctx, `
        SELECT 
            transaction, request_id, currency, provider, amount,
            payment_dt, bank, delivery_cost, goods_total, custom_fee
        FROM payments
        WHERE order_uid = $1`,
		orderUID,
	).Scan(
		&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount,
		&payment.PaymentDt, &payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
	)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	order.Payment = payment

	rows, err := r.db.Query(ctx, `
        SELECT 
            chrt_id, track_number, price, rid, name, sale,
            size, total_price, nm_id, brand, status
        FROM items
        WHERE order_uid = $1`,
		orderUID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query items: %w", err)
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(
			&item.ChrtID, &item.TrackNumber, &item.Price, &item.RID, &item.Name, &item.Sale,
			&item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
		); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		items = append(items, item)
	}
	order.Items = items

	return &order, nil
}

func (r *orderRepository) GetAllOrders(ctx context.Context) ([]models.Order, error) {
	rows, err := r.db.Query(ctx, "SELECT order_uid FROM orders")
	if err != nil {
		return nil, fmt.Errorf("failed to query orders: %w", err)
	}
	defer rows.Close()

	var orderIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan order ID: %w", err)
		}
		orderIDs = append(orderIDs, id)
	}

	var orders []models.Order
	for _, id := range orderIDs {
		order, err := r.GetOrderByID(ctx, id)
		if err != nil {
			r.logger.Error("failed to get order by ID", zap.String("order_uid", id), zap.Error(err))
			continue
		}
		if order != nil {
			orders = append(orders, *order)
		}
	}

	return orders, nil
}
