import requests

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

def test_post_v1_orders_with_valid_authorization_and_payload():
    # First, login to obtain a valid JWT token
    login_url = f"{BASE_URL}/v1/auth/login"
    login_payload = {
        "username": "testuser",
        "password": "TestPassword123!"
    }
    login_headers = {
        "Content-Type": "application/json"
    }
    login_resp = requests.post(login_url, json=login_payload, headers=login_headers, timeout=TIMEOUT)
    assert login_resp.status_code == 200, f"Login failed with status {login_resp.status_code}"
    login_json = login_resp.json()
    assert "token" in login_json and login_json["token"], "No token in login response"
    token = login_json["token"]

    # Create order payload
    order_url = f"{BASE_URL}/v1/orders"
    order_payload = {
        "customer_id": "customer-123",
        "items": [
            {
                "menu_item_id": "menuitem-456",
                "quantity": 2
            }
        ],
        "total_amount": 25.50
    }
    order_headers = {
        "Content-Type": "application/json",
        "Authorization": f"Bearer {token}"
    }

    # Send POST to create order
    order_resp = requests.post(order_url, json=order_payload, headers=order_headers, timeout=TIMEOUT)
    assert order_resp.status_code == 201, f"Order creation failed with status {order_resp.status_code}"
    order_json = order_resp.json()
    assert "order_id" in order_json and order_json["order_id"], "No order_id returned in response"

test_post_v1_orders_with_valid_authorization_and_payload()
