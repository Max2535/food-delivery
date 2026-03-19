import requests
import uuid

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

def test_get_v1_kitchen_status_with_valid_order_id():
    # First, register a new user to ensure the user exists
    register_payload = {
        "username": "testuser",
        "password": "testpassword",
        "email": "testuser@example.com"
    }
    try:
        register_resp = requests.post(f"{BASE_URL}/v1/auth/register", json=register_payload, timeout=TIMEOUT)
        # 201 Created expected or 400/409 if user already exists
        assert register_resp.status_code in {201, 400, 409}, f"Registration failed with status {register_resp.status_code}"

        login_payload = {
            "username": "testuser",
            "password": "testpassword"
        }
        login_resp = requests.post(f"{BASE_URL}/v1/auth/login", json=login_payload, timeout=TIMEOUT)
        assert login_resp.status_code == 200, f"Login failed with status {login_resp.status_code}"
        token = login_resp.json().get("token")
        assert token, "JWT token missing in login response"

        headers = {
            "Authorization": f"Bearer {token}"
        }

        # Create a new order to obtain a valid order_id
        order_payload = {
            "customer_id": str(uuid.uuid4()),
            "items": [
                {
                    "menu_item_id": "item123",
                    "quantity": 1,
                    "portion_multiplier": 1.0
                }
            ],
            "total_amount": 9.99
        }
        order_resp = requests.post(f"{BASE_URL}/v1/orders", json=order_payload, headers=headers, timeout=TIMEOUT)
        assert order_resp.status_code == 201, f"Order creation failed with status {order_resp.status_code}"
        order_id = order_resp.json().get("order_id")
        assert order_id, "order_id missing in order creation response"
        
        try:
            # GET kitchen status for the created order
            kitchen_resp = requests.get(f"{BASE_URL}/v1/kitchen/status/{order_id}", timeout=TIMEOUT)
            assert kitchen_resp.status_code == 200, f"Expected status 200 but got {kitchen_resp.status_code}"
            resp_json = kitchen_resp.json()
            # Validate required fields in response
            assert "status" in resp_json, "'status' field missing in kitchen status response"
            assert resp_json["status"] in {"Received", "Cooking", "Ready"}, "Invalid kitchen ticket status value"
            assert "ticket" in resp_json, "'ticket' field missing in kitchen status response"
            ticket = resp_json["ticket"]

            # ticket details should have items array (if exists)
            assert isinstance(ticket, dict), "'ticket' should be a dictionary"
            # items key can be optional if no items; if exists, should be a list
            if "items" in ticket:
                assert isinstance(ticket["items"], list), "'items' in ticket should be a list"
        finally:
            # Cleanup: delete the created order - assuming DELETE endpoint exists
            # If endpoint does not exist, ignore cleanup or log
            try:
                requests.delete(f"{BASE_URL}/v1/orders/{order_id}", headers=headers, timeout=TIMEOUT)
            except Exception:
                pass

    except requests.RequestException as e:
        assert False, f"Request failed: {str(e)}"

test_get_v1_kitchen_status_with_valid_order_id()
