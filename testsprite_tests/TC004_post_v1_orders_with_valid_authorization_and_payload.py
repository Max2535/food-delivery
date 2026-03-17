import requests

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

def test_post_v1_orders_with_valid_authorization_and_payload():
    login_url = f"{BASE_URL}/v1/auth/login"
    orders_url = f"{BASE_URL}/v1/orders"
    # Valid login credentials (assuming test user exists)
    login_payload = {
        "username": "testuser",
        "password": "TestPass123!"
    }
    try:
        # Step 1: Obtain JWT token via login
        login_response = requests.post(login_url, json=login_payload, timeout=TIMEOUT)
        assert login_response.status_code == 200, f"Expected 200 OK on login, got {login_response.status_code}"
        login_json = login_response.json()
        token = login_json.get("token")
        assert token and isinstance(token, str), "JWT token missing or invalid in login response"

        headers = {
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json"
        }

        # Step 2: Prepare valid order payload
        order_payload = {
            "customer_id": "customer-123",
            "items": [
                {
                    "menu_item_id": "menuitem-abc",
                    "quantity": 2
                },
                {
                    "menu_item_id": "menuitem-def",
                    "quantity": 1
                }
            ],
            "total_amount": 29.99
        }

        # Step 3: Create order
        order_response = requests.post(orders_url, headers=headers, json=order_payload, timeout=TIMEOUT)
        assert order_response.status_code == 201, f"Expected 201 Created for order creation, got {order_response.status_code}"
        order_json = order_response.json()
        order_id = order_json.get("order_id")
        assert order_id and isinstance(order_id, str), "order_id missing or invalid in order creation response"

    finally:
        # Clean up: attempt to delete order if order_id was created
        # Since deletion endpoint is not provided, skipping deletion step
        pass

test_post_v1_orders_with_valid_authorization_and_payload()