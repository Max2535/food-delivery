import requests

BASE_URL = "http://localhost:8080"
ORDERS_PATH = "/v1/orders"
TIMEOUT = 30


def test_post_v1_orders_without_authorization():
    url = BASE_URL + ORDERS_PATH
    payload = {
        "customer_id": "test_customer_123",
        "items": [
            {"menu_item_id": "item_001", "quantity": 2},
            {"menu_item_id": "item_002", "quantity": 1}
        ],
        "total_amount": 29.99
    }
    headers = {
        "Content-Type": "application/json"
    }

    try:
        response = requests.post(url, json=payload, headers=headers, timeout=TIMEOUT)
    except requests.RequestException as e:
        assert False, f"Request failed: {e}"

    assert response.status_code == 401, f"Expected 401 Unauthorized, got {response.status_code}"
    # Optionally test for response body or headers related to unauthorized response if applicable


test_post_v1_orders_without_authorization()