import requests

BASE_URL = "http://localhost:8080"
ORDERS_PATH = "/v1/orders"
TIMEOUT = 30

def test_post_v1_orders_without_authorization_header():
    url = BASE_URL + ORDERS_PATH
    headers = {
        "Content-Type": "application/json"
    }
    # Minimal valid order payload based on PRD (customer_id, items, total_amount required)
    payload = {
        "customer_id": "test-customer-123",
        "items": [
            {
                "menu_item_id": "item-001",
                "quantity": 1
            }
        ],
        "total_amount": 9.99
    }

    try:
        response = requests.post(url, json=payload, headers=headers, timeout=TIMEOUT)
    except requests.RequestException as e:
        assert False, f"Request failed: {e}"

    # Expect 401 Unauthorized
    assert response.status_code == 401, f"Expected 401 Unauthorized but got {response.status_code}"

test_post_v1_orders_without_authorization_header()