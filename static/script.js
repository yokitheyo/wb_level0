function searchOrder() {
    const orderUID = document.getElementById('orderUID').value.trim();
    
    if (!orderUID) {
        showError('Пожалуйста, введите ID заказа');
        return;
    }

    hideError();
    hideOrderResult();
    showLoading();

    fetch(`/order/${orderUID}`)
        .then(response => {
            if (!response.ok) {
                if (response.status === 404) {
                    throw new Error('Заказ не найден');
                } else if (response.status === 400) {
                    throw new Error('Неверный формат ID заказа');
                } else {
                    throw new Error('Ошибка сервера');
                }
            }
            return response.json();
        })
        .then(order => {
            hideLoading();
            displayOrder(order);
        })
        .catch(error => {
            hideLoading();
            showError(error.message);
        });
}

function displayOrder(order) {
    const orderData = document.getElementById('orderData');
    
    orderData.innerHTML = `
        <div class="order-info">
            <div class="info-section">
                <h3>Основная информация</h3>
                <div class="info-row">
                    <span class="info-label">ID заказа:</span>
                    <span class="info-value">${order.order_uid}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Трек-номер:</span>
                    <span class="info-value">${order.track_number}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Точка входа:</span>
                    <span class="info-value">${order.entry}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Локаль:</span>
                    <span class="info-value">${order.locale}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">ID клиента:</span>
                    <span class="info-value">${order.customer_id}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Служба доставки:</span>
                    <span class="info-value">${order.delivery_service}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Дата создания:</span>
                    <span class="info-value">${formatDate(order.date_created)}</span>
                </div>
            </div>

            <div class="info-section">
                <h3>Доставка</h3>
                <div class="info-row">
                    <span class="info-label">Имя:</span>
                    <span class="info-value">${order.delivery.name}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Телефон:</span>
                    <span class="info-value">${order.delivery.phone}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Индекс:</span>
                    <span class="info-value">${order.delivery.zip}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Город:</span>
                    <span class="info-value">${order.delivery.city}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Адрес:</span>
                    <span class="info-value">${order.delivery.address}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Регион:</span>
                    <span class="info-value">${order.delivery.region}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Email:</span>
                    <span class="info-value">${order.delivery.email}</span>
                </div>
            </div>

            <div class="info-section">
                <h3>Оплата</h3>
                <div class="info-row">
                    <span class="info-label">Транзакция:</span>
                    <span class="info-value">${order.payment.transaction}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Валюта:</span>
                    <span class="info-value">${order.payment.currency}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Провайдер:</span>
                    <span class="info-value">${order.payment.provider}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Сумма:</span>
                    <span class="info-value">${order.payment.amount}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Банк:</span>
                    <span class="info-value">${order.payment.bank}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Стоимость доставки:</span>
                    <span class="info-value">${order.payment.delivery_cost}</span>
                </div>
                <div class="info-row">
                    <span class="info-label">Стоимость товаров:</span>
                    <span class="info-value">${order.payment.goods_total}</span>
                </div>
            </div>
        </div>

        <div class="items-section">
            <h3>Товары (${order.items.length})</h3>
            <div class="items-grid">
                ${order.items.map(item => `
                    <div class="item-card">
                        <h4>${item.name}</h4>
                        <div class="info-row">
                            <span class="info-label">ID товара:</span>
                            <span class="info-value">${item.chrt_id}</span>
                        </div>
                        <div class="info-row">
                            <span class="info-label">Бренд:</span>
                            <span class="info-value">${item.brand}</span>
                        </div>
                        <div class="info-row">
                            <span class="info-label">Размер:</span>
                            <span class="info-value">${item.size}</span>
                        </div>
                        <div class="info-row">
                            <span class="info-label">Цена:</span>
                            <span class="info-value">${item.price}</span>
                        </div>
                        <div class="info-row">
                            <span class="info-label">Скидка:</span>
                            <span class="info-value">${item.sale}%</span>
                        </div>
                        <div class="info-row">
                            <span class="info-label">Итоговая цена:</span>
                            <span class="info-value">${item.total_price}</span>
                        </div>
                        <div class="info-row">
                            <span class="info-label">Статус:</span>
                            <span class="info-value">${item.status}</span>
                        </div>
                    </div>
                `).join('')}
            </div>
        </div>
    `;

    showOrderResult();
}

function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('ru-RU');
}

function showLoading() {
    document.getElementById('loading').style.display = 'block';
}

function hideLoading() {
    document.getElementById('loading').style.display = 'none';
}

function showError(message) {
    const errorDiv = document.getElementById('error');
    errorDiv.textContent = message;
    errorDiv.style.display = 'block';
}

function hideError() {
    document.getElementById('error').style.display = 'none';
}

function showOrderResult() {
    document.getElementById('orderResult').style.display = 'block';
}

function hideOrderResult() {
    document.getElementById('orderResult').style.display = 'none';
}

document.getElementById('orderUID').addEventListener('keypress', function(e) {
    if (e.key === 'Enter') {
        searchOrder();
    }
});
