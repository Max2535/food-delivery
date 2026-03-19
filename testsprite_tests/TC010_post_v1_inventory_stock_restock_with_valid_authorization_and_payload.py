import requests

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

def test_post_v1_inventory_stock_restock_with_valid_authorization_and_payload():
    login_url = f"{BASE_URL}/v1/auth/login"
    restock_url = f"{BASE_URL}/v1/inventory/stock/restock"
    materials_url = f"{BASE_URL}/v1/inventory/materials"

    # Credentials for login - adjust if needed
    login_payload = {
        "username": "testuser",
        "password": "TestPass123!"
    }

    # Authenticate and get JWT token
    try:
        login_response = requests.post(login_url, json=login_payload, timeout=TIMEOUT)
        assert login_response.status_code == 200, f"Login failed with status {login_response.status_code}"
        token = login_response.json().get("token")
        assert token, "No token received in login response"
    except Exception as e:
        raise AssertionError(f"Authentication failed: {e}")

    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }

    # Get material_id to restock; if none exist, test must fail
    try:
        materials_response = requests.get(materials_url, timeout=TIMEOUT)
        assert materials_response.status_code == 200, f"Failed to get materials: {materials_response.status_code}"
        resp_json = materials_response.json()
        materials = resp_json.get("raw_materials", [])
        assert isinstance(materials, list) and len(materials) > 0, "No materials found for restock"
        # Pick the first material with non-null id
        material_id = None
        for m in materials:
            if "material_id" in m:
                material_id = m["material_id"]
                break
            if "id" in m:
                material_id = m["id"]
                break
        assert material_id, "No valid material_id found in materials"
    except Exception as e:
        raise AssertionError(f"Failed to obtain material_id for restock: {e}")

    restock_payload = {
        "material_id": material_id,
        "quantity": 10,
        "note": "Test restock via automated test"
    }

    # Call restock endpoint
    try:
        restock_response = requests.post(restock_url, headers=headers, json=restock_payload, timeout=TIMEOUT)
        assert restock_response.status_code == 201 or restock_response.status_code == 200, f"Unexpected status code: {restock_response.status_code}"
        json_resp = restock_response.json()
        assert "transaction_id" in json_resp and json_resp["transaction_id"], "transaction_id missing or empty in response"
    except Exception as e:
        raise AssertionError(f"Restock request failed or invalid response: {e}")

test_post_v1_inventory_stock_restock_with_valid_authorization_and_payload()