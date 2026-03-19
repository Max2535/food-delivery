import requests
import uuid

BASE_URL = "http://localhost:8080"
TIMEOUT = 30
LOGIN_ENDPOINT = f"{BASE_URL}/v1/auth/login"
ORDER_ENDPOINT = f"{BASE_URL}/v1/orders"
KITCHEN_STATUS_ENDPOINT = f"{BASE_URL}/v1/kitchen/status"

VALID_LOGIN_PAYLOAD = {
    "username": "testuser",
    "password": "TestPass123!"
}

def test_get_v1_kitchen_status_with_valid_order_id():
    # Step 1: Login to get JWT token
    try:
        login_resp = requests.post(
            LOGIN_ENDPOINT,
            json=VALID_LOGIN_PAYLOAD,
            timeout=TIMEOUT
        )
        assert login_resp.status_code == 200, f"Login failed: {login_resp.text}"
        token = login_resp.json().get("token")
        assert token, "JWT token not found in login response"
        headers_auth = {"Authorization": f"Bearer {token}"}
    except Exception as e:
        raise AssertionError(f"Authentication step failed: {e}")

    # Step 2: Create order to get valid order_id
    order_payload = {
        "customer_id": str(uuid.uuid4()),
        "items": [
            {
                "menu_item_id": "item1",
                "quantity": 1,
                "portion_multiplier": 1.0
            }
        ],
        "total_amount": 19.99
    }

    order_id = None
    try:
        create_order_resp = requests.post(
            ORDER_ENDPOINT,
            headers=headers_auth,
            json=order_payload,
            timeout=TIMEOUT
        )
        assert create_order_resp.status_code == 201, f"Order creation failed: {create_order_resp.text}"
        order_id = create_order_resp.json().get("id")
        assert order_id, "order_id not found in order creation response"

        # Step 3: Get kitchen status with the valid order_id
        kitchen_status_url = f"{KITCHEN_STATUS_ENDPOINT}/{str(order_id)}"
        kitchen_resp = requests.get(kitchen_status_url, timeout=TIMEOUT)

        assert kitchen_resp.status_code == 200, f"Expected 200 OK, got {kitchen_resp.status_code}"

        kitchen_data = kitchen_resp.json()
        # Validate presence of kitchen ticket status and ticket details
        assert "status" in kitchen_data, "Kitchen status missing in response"
        assert kitchen_data["status"] in ["Received", "Cooking", "Ready"], "Invalid kitchen status value"
        assert "ticket" in kitchen_data, "Kitchen ticket details missing in response"
        ticket = kitchen_data["ticket"]
        assert isinstance(ticket, dict), "Ticket details should be a dictionary"
        assert "items" in ticket and isinstance(ticket["items"], list), "Ticket items missing or invalid type"

    finally:
        # Clean up: delete the order if deletion endpoint exists (Not specified in PRD)
        pass

test_get_v1_kitchen_status_with_valid_order_id()
