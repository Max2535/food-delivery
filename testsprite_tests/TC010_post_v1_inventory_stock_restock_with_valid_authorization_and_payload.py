import requests

BASE_URL = "http://localhost:8080"
AUTH_TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0dXNlciIsImlhdCI6MTY4OTM2MDYwMCwiZXhwIjoxNjg5MzkxNjAwfQ.XDJa7RzXaVZ-6_y1-5Qm--lw3S_v76zTRPlnVITsPJg"  # Placeholder JWT; replace with valid token for real tests
TIMEOUT = 30

def test_post_v1_inventory_stock_restock_with_valid_authorization_and_payload():
    headers = {
        "Authorization": f"Bearer {AUTH_TOKEN}",
        "Content-Type": "application/json"
    }
    # Step 1: Get materials to find a valid material_id
    try:
        response = requests.get(f"{BASE_URL}/v1/inventory/materials", timeout=TIMEOUT)
        response.raise_for_status()
        materials = response.json()
        assert isinstance(materials, list) and len(materials) > 0, "No materials found to restock."
        material_id = None
        for m in materials:
            if "material_id" in m:
                material_id = m["material_id"]
                break
        assert material_id is not None, "Valid material_id not found in materials list."
    except Exception as e:
        raise AssertionError(f"Failed to retrieve materials or find valid material_id: {e}")

    payload = {
        "material_id": material_id,
        "quantity": 10,
        "note": "Restocking for test case TC010"
    }

    try:
        restock_response = requests.post(
            f"{BASE_URL}/v1/inventory/stock/restock",
            headers=headers,
            json=payload,
            timeout=TIMEOUT
        )
        assert restock_response.status_code == 200 or restock_response.status_code == 201, \
            f"Expected 200 or 201 status but got {restock_response.status_code}"
        restock_data = restock_response.json()
        assert "transaction_id" in restock_data and restock_data["transaction_id"], "transaction_id not present in response"
    except requests.exceptions.RequestException as e:
        raise AssertionError(f"HTTP request failed: {e}")

test_post_v1_inventory_stock_restock_with_valid_authorization_and_payload()