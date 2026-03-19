import requests

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

def test_post_v1_catalog_menus_with_invalid_bom_entry():
    token = get_jwt_token()
    url = f"{BASE_URL}/v1/catalog/menus"
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    
    # Define invalid BOM payloads where ingredient_id and sub_menu_item_id are both set or both null.
    invalid_bom_entries = [
        {
            # Both ingredient_id and sub_menu_item_id set
            "name": "Test Menu Invalid BOM Both Set",
            "price": 9.99,
            "category": "Test Category",
            "availability": True,
            "bom": [
                {
                    "ingredient_id": "ing-123",
                    "sub_menu_item_id": "sub-456",
                    "quantity": 2
                }
            ]
        },
        {
            # Both ingredient_id and sub_menu_item_id null
            "name": "Test Menu Invalid BOM Both Null",
            "price": 12.99,
            "category": "Test Category",
            "availability": True,
            "bom": [
                {
                    "ingredient_id": None,
                    "sub_menu_item_id": None,
                    "quantity": 1
                }
            ]
        }
    ]

    for payload in invalid_bom_entries:
        try:
            response = requests.post(url, headers=headers, json=payload, timeout=30)
        except requests.RequestException as e:
            assert False, f"Request failed with exception: {e}"
        if response.status_code == 401:
            assert False, "Unauthorized: Please provide a valid JWT token in AUTH_TOKEN to run this test."
        # Expecting HTTP 400 Bad Request for invalid BOM entries
        assert response.status_code == 400, f"Expected 400 Bad Request but got {response.status_code}. Response: {response.text}"

test_post_v1_catalog_menus_with_invalid_bom_entry()
