document.addEventListener('DOMContentLoaded', () => {
    const orderUidInput = document.getElementById('order-uid');
    const searchBtn = document.getElementById('search-btn');
    const orderInfo = document.getElementById('order-info');
    const itemsSection = document.getElementById('items-section');
    const orderDetails = document.querySelector('.order-details');
    const itemsGrid = document.querySelector('.items-grid');
    const prevBtn = document.getElementById('prev-btn');
    const nextBtn = document.getElementById('next-btn');
    const errorMessage = document.getElementById('error-message');

    let currentOrderUid = '';
    let lastItemId = 0;
    let paginationHistory = [];
    const ITEMS_PER_PAGE = 4;

    searchBtn.addEventListener('click', () => {
        const orderUid = orderUidInput.value.trim();
        if (!orderUid) {
            showError('Please enter an Order UID');
            return;
        }
        
        resetState();
        currentOrderUid = orderUid;
        paginationHistory = [0];
        lastItemId = 0;
        fetchOrder(orderUid);
    });

    prevBtn.addEventListener('click', () => {
        if (paginationHistory.length > 1) {
            paginationHistory.pop();
            lastItemId = paginationHistory[paginationHistory.length - 1];
            fetchItems(currentOrderUid, lastItemId, ITEMS_PER_PAGE);
        }
    });

    nextBtn.addEventListener('click', () => {
        fetchItems(currentOrderUid, lastItemId, ITEMS_PER_PAGE);
    });

    function fetchOrder(orderUid) {
        showLoading(orderDetails);
        
        fetch(`/api/orders/${orderUid}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error(`Order not found (status: ${response.status})`);
                }
                return response.json();
            })
            .then(order => {
                displayOrder(order);
                fetchItems(orderUid, 0, ITEMS_PER_PAGE);
            })
            .catch(error => {
                showError(error.message);
                hideLoading(orderDetails);
            });
    }

    function fetchItems(orderUid, lastId, limit) {
        showLoading(itemsGrid);
        
        fetch(`/api/orders/${orderUid}/items?last_id=${lastId}&limit=${limit}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error(`Failed to load items (status: ${response.status})`);
                }
                return response.json();
            })
            .then(items => {
                displayItems(items, lastId);
                updatePaginationButtons(items.length === ITEMS_PER_PAGE);
            })
            .catch(error => {
                showError(error.message);
                hideLoading(itemsGrid);
            });
    }

    function displayOrder(order) {
        orderInfo.classList.remove('hidden');
        
        const orderHTML = `
            <div class="order-card">
                <p><strong>Order UID:</strong> ${order.order_uid}</p>
                <p><strong>Track Number:</strong> ${order.track_number}</p>
                <p><strong>Entry:</strong> ${order.entry}</p>
                <p><strong>Locale:</strong> ${order.locale}</p>
                <p><strong>Customer ID:</strong> ${order.customer_id}</p>
                <p><strong>Date Created:</strong> ${new Date(order.date_created).toLocaleString()}</p>
            </div>
        `;
        
        orderDetails.innerHTML = orderHTML;
        hideLoading(orderDetails);
    }

    function displayItems(items, lastId) {
        itemsSection.classList.remove('hidden');
        
        if (items.length === 0) {
            itemsGrid.innerHTML = '<p>No items found for this order</p>';
            hideLoading(itemsGrid);
            return;
        }
        
        if (lastId !== 0 || paginationHistory.length === 1) {
            paginationHistory.push(lastId);
        }
        
        if (items.length > 0) {
            lastItemId = items[items.length - 1].id;
        }
        
        let itemsHTML = '';
        items.forEach(item => {
            itemsHTML += `
                <div class="item-card">
                    <p><strong>Name:</strong> ${item.name}</p>
                    <p><strong>Price:</strong> ${item.price}</p>
                    <p><strong>Size:</strong> ${item.size}</p>
                    <p><strong>Brand:</strong> ${item.brand}</p>
                </div>
            `;
        });
        
        itemsGrid.innerHTML = itemsHTML;
        hideLoading(itemsGrid);
    }

    function updatePaginationButtons(hasMoreItems) {
        prevBtn.disabled = paginationHistory.length <= 1;
        nextBtn.disabled = !hasMoreItems;
    }

    function showLoading(element) {
        element.innerHTML = '<p>Loading...</p>';
    }

    function hideLoading(element) {
        if (element.innerHTML.includes('Loading')) {
            element.innerHTML = '';
        }
    }

    function showError(message) {
        errorMessage.textContent = message;
        errorMessage.classList.remove('hidden');
        
        setTimeout(() => {
            errorMessage.classList.add('hidden');
        }, 5000);
    }

    function resetState() {
        orderInfo.classList.add('hidden');
        itemsSection.classList.add('hidden');
        orderDetails.innerHTML = '';
        itemsGrid.innerHTML = '';
        currentOrderUid = '';
        lastItemId = 0;
        paginationHistory = [];
        prevBtn.disabled = true;
        nextBtn.disabled = true;
        errorMessage.classList.add('hidden');
    }
});