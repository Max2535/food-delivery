import requests
import uuid

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

def get_jwt_token():
    # For this test, we assume we have valid credentials for an existing user.
    login_url = f"{BASE_URL}/v1/auth/login"
    login_payload = {
        "username": "testuser",
        "password": "TestPass123!"
    }
    try:
        resp = requests.post(login_url, json=login_payload, timeout=TIMEOUT)
        assert resp.status_code == 200, f"Login failed with status {resp.status_code}"
        data = resp.json()
        token = data.get("token") or data.get("jwt") or data.get("access_token")
        assert token, "JWT token not found in login response"
        return token
    except Exception as e:
        raise RuntimeError(f"Could not obtain JWT token: {e}")

def test_post_v1_catalog_menus_with_valid_authorization_and_payload():
    token = get_jwt_token()
    url = f"{BASE_URL}/v1/catalog/menus"
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json",
    }

    # Construct a valid menu payload
    # Example fields according to typical menu schema described in overview (name, price, category, availability)
    # Using unique name to avoid collisions.
    unique_name = f"Test Menu {uuid.uuid4()}"
    payload = {
        "name": unique_name,
        "price": 9.99,
        "category": "Test Category",
        "availability": True,
        # BOM or ingredients are not explicitly required, so omit or provide empty list.
        "bom": [],  # Assuming optional, if required fill accordingly.
        "description": "Automated test menu item"
    }

    menu_id = None
    try:
        response = requests.post(url, headers=headers, json=payload, timeout=TIMEOUT)
        assert response.status_code == 201, f"Expected 201 Created but got {response.status_code}"
        resp_json = response.json()
        menu_id = resp_json.get("menu_id") or resp_json.get("id")
        assert menu_id, "Response missing menu_id"
    finally:
        # Cleanup - delete the created menu if created
        if menu_id:
            delete_url = f"{BASE_URL}/v1/catalog/menus/{menu_id}"
            try:
                del_resp = requests.delete(delete_url, headers=headers, timeout=TIMEOUT)
                # Accept 200 OK or 204 No Content as successful deletion
                assert del_resp.status_code in (200, 204), f"Failed to delete created menu with status {del_resp.status_code}"
            except Exception:
                # Log deletion failure but do not raise to avoid masking original test result
                pass

test_post_v1_catalog_menus_with_valid_authorization_and_payload()