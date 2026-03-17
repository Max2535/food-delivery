import requests
import uuid
import time

BASE_URL = "http://localhost:8080"
AUTH_URL = f"{BASE_URL}/v1/auth/login"
ORDER_URL = f"{BASE_URL}/v1/orders"
KITCHEN_STATUS_URL = f"{BASE_URL}/v1/kitchen/status"

USERNAME = "testuser"
PASSWORD = "TestPass123!"

def test_get_v1_kitchen_status_with_valid_order_id():
    timeout = 30
    # Step 1: Authenticate to get JWT token
    login_payload = {
        "username": USERNAME,
        "password": PASSWORD
    }
    token = None
    order_id = None
    headers = {"Content-Type": "application/json"}
    try:
        login_resp = requests.post(AUTH_URL, json=login_payload, headers=headers, timeout=timeout)
        assert login_resp.status_code == 200, f"Login failed with status {login_resp.status_code}"
        token = login_resp.json().get("token")
        assert token, "JWT token not found in login response"

        auth_headers = {
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json"
        }

        # Step 2: Create a new order with valid payload
        order_payload = {
            "customer_id": str(uuid.uuid4()),
            "items": [
                {"menu_item_id": str(uuid.uuid4()), "quantity": 1, "portion_multiplier": 1.0}
            ],
            "total_amount": 10.0
        }

        order_resp = requests.post(ORDER_URL, json=order_payload, headers=auth_headers, timeout=timeout)
        assert order_resp.status_code == 201, f"Order creation failed with status {order_resp.status_code}"
        order_resp_json = order_resp.json()
        order_id = order_resp_json.get("order_id")
        assert order_id, "order_id not found in order creation response"

        # Allow some time for kitchen ticket to be created
        time.sleep(2)

        # Step 3: Get kitchen ticket status by order_id
        kitchen_status_resp = requests.get(f"{KITCHEN_STATUS_URL}/{order_id}", timeout=timeout)
        assert kitchen_status_resp.status_code == 200, f"Kitchen status fetch failed with status {kitchen_status_resp.status_code}"
        kitchen_status_json = kitchen_status_resp.json()

        # Validate kitchen ticket status and ticket details presence
        assert "status" in kitchen_status_json, "'status' key missing in kitchen status response"
        assert kitchen_status_json["status"] in ["Received", "Cooking", "Ready"], "Invalid kitchen status value"
        assert "ticket" in kitchen_status_json, "'ticket' key missing in kitchen status response"
        assert isinstance(kitchen_status_json["ticket"], dict), "Ticket details should be a dictionary"

    finally:
        # Cleanup: delete the created order if possible
        if token and order_id:
            # No explicit delete order endpoint described in PRD, skipping deletion
            pass

test_get_v1_kitchen_status_with_valid_order_id()