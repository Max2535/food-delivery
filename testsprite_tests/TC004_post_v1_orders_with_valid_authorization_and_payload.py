import requests

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

def test_post_v1_orders_with_valid_authorization_and_payload():
    # First, login to get a valid JWT token
    login_url = f"{BASE_URL}/v1/auth/login"
    login_payload = {
        "username": "testuser",
        "password": "testpassword"
    }
    login_headers = {"Content-Type": "application/json"}
    try:
        resp_login = requests.post(login_url, json=login_payload, headers=login_headers, timeout=TIMEOUT)
        assert resp_login.status_code == 200, f"Login failed with status {resp_login.status_code}"
        json_login = resp_login.json()
        assert "token" in json_login or "access_token" in json_login, "JWT token not found in login response"
        token = json_login.get("token") or json_login.get("access_token")
    except (requests.RequestException, AssertionError) as e:
        raise AssertionError(f"Login request or validation failed: {e}")

    headers_order = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {token}"
    }

    # Prepare a valid order payload with dummy data
    # Since PRD does not specify exact fields for items, assume items as list of dict with menu_item_id and quantity
    order_payload = {
        "customer_id": "cust_12345",
        "items": [
            {"menu_item_id": "menu_001", "quantity": 2},
            {"menu_item_id": "menu_002", "quantity": 1}
        ],
        "total_amount": 29.99
    }

    order_url = f"{BASE_URL}/v1/orders"

    try:
        response = requests.post(order_url, json=order_payload, headers=headers_order, timeout=TIMEOUT)
        assert response.status_code == 201, f"Expected 201 Created, got {response.status_code}"
        json_resp = response.json()
        assert "order_id" in json_resp, "order_id missing in response"
    except (requests.RequestException, AssertionError) as e:
        raise AssertionError(f"Order creation request or validation failed: {e}")

test_post_v1_orders_with_valid_authorization_and_payload()