import requests

BASE_URL = "http://localhost:8080"
LOGIN_ENDPOINT = "/v1/auth/login"
MATERIALS_ENDPOINT = "/v1/inventory/materials"
RESTOCK_ENDPOINT = "/v1/inventory/stock/restock"

USERNAME = "testuser"
PASSWORD = "testpassword"

def test_post_v1_inventory_stock_restock_with_valid_authorization_and_payload():
    try:
        # Step 1: Login to get JWT token
        login_payload = {"username": USERNAME, "password": PASSWORD}
        login_response = requests.post(
            BASE_URL + LOGIN_ENDPOINT,
            json=login_payload,
            timeout=30
        )
        assert login_response.status_code == 200, f"Login failed with status {login_response.status_code}"
        token = login_response.json().get("token")
        assert token, "No token found in login response"

        headers = {
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json"
        }

        # Step 2: Retrieve list of materials to get a valid material_id
        materials_response = requests.get(
            BASE_URL + MATERIALS_ENDPOINT,
            timeout=30
        )
        assert materials_response.status_code == 200, f"Failed to get materials with status {materials_response.status_code}"
        materials = materials_response.json()
        assert isinstance(materials, list) and len(materials) > 0, "No materials available to restock"
        material_id = materials[0].get("material_id") or materials[0].get("id")
        assert material_id, "Material ID not found in materials response"

        # Step 3: Prepare restock payload
        restock_payload = {
            "material_id": material_id,
            "quantity": 10,
            "note": "Test restock via automated test"
        }

        # Step 4: Send POST request to restock endpoint
        restock_response = requests.post(
            BASE_URL + RESTOCK_ENDPOINT,
            headers=headers,
            json=restock_payload,
            timeout=30
        )
        # According to PRD, expected code is 200 Created - assuming status code 200 
        assert restock_response.status_code == 200, f"Restock failed with status {restock_response.status_code}"
        response_json = restock_response.json()
        assert "transaction_id" in response_json and response_json["transaction_id"], "transaction_id missing in restock response"

    except requests.RequestException as e:
        assert False, f"RequestException occurred: {e}"

test_post_v1_inventory_stock_restock_with_valid_authorization_and_payload()