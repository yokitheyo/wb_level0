-- Индексы для улучшения производительности

-- Индекс по дате создания для быстрой сортировки при восстановлении кеша
CREATE INDEX IF NOT EXISTS idx_orders_date_created ON orders(date_created DESC);

-- Индекс по track_number для поиска
CREATE INDEX IF NOT EXISTS idx_orders_track_number ON orders(track_number);

-- Индекс по customer_id для поиска заказов клиента
CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);

-- Индексы для связанных таблиц (хотя они уже есть через FK)
CREATE INDEX IF NOT EXISTS idx_deliveries_order_uid ON deliveries(order_uid);
CREATE INDEX IF NOT EXISTS idx_payments_order_uid ON payments(order_uid);
CREATE INDEX IF NOT EXISTS idx_items_order_uid ON items(order_uid);

-- Индекс для поиска товаров по характеристикам
CREATE INDEX IF NOT EXISTS idx_items_chrt_id ON items(chrt_id);
CREATE INDEX IF NOT EXISTS idx_items_nm_id ON items(nm_id);
