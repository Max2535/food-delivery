import requests

BASE_URL = "http://localhost:8080"

def test_post_v1_orders_without_authorization_header():
    url = f"{BASE_URL}/v1/orders"
    payload = {
        "customer_id": "test-customer-id",
        "items": [
            {
                "menu_item_id": "test-menu-item-id",
                "quantity": 1,
                "portion_multiplier": 1.0
            }
        ],
        "total_amount": 15.50
    }
    headers = {
        "Content-Type": "application/json"
        # No Authorization header intentionally
    }
    try:
        response = requests.post(url, json=payload, headers=headers, timeout=30)
    except requests.RequestException as e:
        assert False, f"Request failed: {e}"
    assert response.status_code == 401, f"Expected 401 Unauthorized but got {response.status_code}"
    # Optional: check response content if any standard error message expected
    # e.g. assert "Unauthorized" in response.text or response.json().get("error") == "Unauthorized"

test_post_v1_orders_without_authorization_header()