import requests

BASE_URL = "http://localhost:8080"
TIMEOUT = 30


def test_post_v1_orders_with_valid_authorization_and_payload():
    # First, login to obtain a valid JWT token
    login_payload = {
        "username": "testuser",
        "password": "testpass"
    }

    login_response = requests.post(
        f"{BASE_URL}/v1/auth/login",
        json=login_payload,
        timeout=TIMEOUT
    )
    assert login_response.status_code == 200, f"Login failed with status code {login_response.status_code}"
    json_login = login_response.json()
    assert "token" in json_login, "Login response JSON does not contain 'token'"

    jwt_token = json_login["token"]
    headers = {
        "Authorization": f"Bearer {jwt_token}",
        "Content-Type": "application/json"
    }

    order_payload = {
        "customer_id": "cust-12345",
        "items": [
            {
                "menu_item_id": "menu-67890",
                "quantity": 2
            },
            {
                "menu_item_id": "menu-13579",
                "quantity": 1
            }
        ],
        "total_amount": 29.97
    }

    response = requests.post(
        f"{BASE_URL}/v1/orders",
        json=order_payload,
        headers=headers,
        timeout=TIMEOUT
    )

    assert response.status_code == 201, f"Expected 201 Created, got {response.status_code}"
    json_response = response.json()
    assert "order_id" in json_response, "Response JSON does not contain 'order_id'"
    order_id = json_response["order_id"]
    assert isinstance(order_id, str) and len(order_id) > 0, "Invalid order_id value"


test_post_v1_orders_with_valid_authorization_and_payload()
