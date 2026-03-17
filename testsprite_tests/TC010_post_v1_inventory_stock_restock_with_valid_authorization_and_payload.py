import requests
import uuid

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

def get_jwt_token():
    login_url = f"{BASE_URL}/v1/auth/login"
    credentials = {
        "username": "testuser",
        "password": "TestPass123!"
    }
    try:
        resp = requests.post(login_url, json=credentials, timeout=TIMEOUT)
        resp.raise_for_status()
        data = resp.json()
        token = data.get("token") or data.get("access_token")
        assert token, "No token found in login response"
        return token
    except Exception as e:
        raise RuntimeError(f"Failed to get JWT token: {e}")

def get_material_id():
    materials_url = f"{BASE_URL}/v1/inventory/materials"
    try:
        resp = requests.get(materials_url, timeout=TIMEOUT)
        resp.raise_for_status()
        data = resp.json()
        assert isinstance(data, list), "Materials response is not a list"
        if not data:
            raise RuntimeError("No materials found to restock")
        # Pick the first material
        material = data[0]
        material_id = material.get("material_id") or material.get("id")
        assert material_id, "Material ID not found in material item"
        return material_id
    except Exception as e:
        raise RuntimeError(f"Failed to get material_id: {e}")

def test_post_v1_inventory_stock_restock_with_valid_authorization_and_payload():
    token = get_jwt_token()
    material_id = get_material_id()

    restock_url = f"{BASE_URL}/v1/inventory/stock/restock"
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    payload = {
        "material_id": material_id,
        "quantity": 10,
        "note": "Restock from automated test"
    }

    try:
        response = requests.post(restock_url, headers=headers, json=payload, timeout=TIMEOUT)
        assert response.status_code == 200, f"Expected 200 Created, got {response.status_code}"
        resp_json = response.json()
        transaction_id = resp_json.get("transaction_id")
        assert transaction_id, "transaction_id not found in response"
    except requests.exceptions.RequestException as e:
        raise AssertionError(f"Request failed: {e}")

test_post_v1_inventory_stock_restock_with_valid_authorization_and_payload()