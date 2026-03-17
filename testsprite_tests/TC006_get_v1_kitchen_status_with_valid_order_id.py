import requests
import uuid

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

# Using a placeholder JWT token since no credentials given to fetch an actual one
# For real testing, this token should be obtained via login and securely passed
# According to PRD, GET /v1/kitchen/status/{order_id} does not require auth.
# But per instructions, authType is Bearer token globally, so we send it if needed.
# Here we omit Authorization as per PRD endpoint spec (auth_required: false).
# If needed uncomment below and provide a valid token.
# AUTH_TOKEN = "Bearer <your-valid-jwt-token>"

def test_get_v1_kitchen_status_with_valid_order_id():
    # Step 1: Create an order to get a valid order_id
    # Need to authenticate first to create an order (POST /v1/auth/login)
    # We'll do a login with a test user (credentials should be replaced with valid ones)
    login_url = f"{BASE_URL}/v1/auth/login"
    login_payload = {
        "username": "testuser",
        "password": "testpassword"
    }
    try:
        login_resp = requests.post(login_url, json=login_payload, timeout=TIMEOUT)
        login_resp.raise_for_status()
        jwt_token = login_resp.json().get("token")
        assert jwt_token, "Login response missing token"
    except Exception as e:
        raise AssertionError(f"Failed to login to get token: {e}")

    headers_auth = {
        "Authorization": f"Bearer {jwt_token}",
        "Content-Type": "application/json"
    }

    # Create an order
    orders_url = f"{BASE_URL}/v1/orders"
    order_payload = {
        "customer_id": str(uuid.uuid4()),
        "items": [
            {
                "menu_item_id": "sample-menu-item-id",
                "quantity": 1,
                "portion_multiplier": 1.0
            }
        ],
        "total_amount": 9.99
    }

    order_id = None
    try:
        order_resp = requests.post(orders_url, headers=headers_auth, json=order_payload, timeout=TIMEOUT)
        order_resp.raise_for_status()
        assert order_resp.status_code == 201, f"Expected 201 Created but got {order_resp.status_code}"
        order_json = order_resp.json()
        order_id = order_json.get("order_id")
        assert order_id, "Response missing order_id"

        # Now test GET /v1/kitchen/status/{order_id}
        # Per PRD, no auth required for this endpoint
        kitchen_status_url = f"{BASE_URL}/v1/kitchen/status/{order_id}"
        kitchen_resp = requests.get(kitchen_status_url, timeout=TIMEOUT)
        kitchen_resp.raise_for_status()
        assert kitchen_resp.status_code == 200, f"Expected 200 OK but got {kitchen_resp.status_code}"
        kitchen_json = kitchen_resp.json()
        # Validate kitchen ticket status presence and value
        status = kitchen_json.get("status")
        assert status in {"Received", "Cooking", "Ready"}, f"Unexpected kitchen status '{status}'"
        # Validate ticket details presence
        ticket = kitchen_json.get("ticket")
        assert ticket is not None, "Missing ticket details"
        # Validate ticket has items array
        items = ticket.get("items")
        assert isinstance(items, list), "Ticket items should be a list"
    finally:
        # Cleanup - delete the created order if an endpoint existed; since there's no order delete endpoint in PRD, skip
        # If needed to clean, implement delete here
        pass

test_get_v1_kitchen_status_with_valid_order_id()