import requests

BASE_URL = "http://localhost:8080"
LOGIN_ENDPOINT = "/v1/auth/login"
MENU_ENDPOINT = "/v1/catalog/menus"

USERNAME = "testuser"
PASSWORD = "testpassword"

def test_post_v1_catalog_menus_with_invalid_bom_entry():
    # Obtain JWT token
    login_payload = {"username": USERNAME, "password": PASSWORD}
    try:
        login_resp = requests.post(
            BASE_URL + LOGIN_ENDPOINT,
            json=login_payload,
            timeout=30
        )
        assert login_resp.status_code == 200, f"Login failed with status code {login_resp.status_code}"
        token = login_resp.json().get("token")
        assert token, "JWT token not found in login response"
    except Exception as e:
        raise AssertionError(f"Failed to login: {e}")

    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }

    # Prepare invalid BOM entries to test all cases where both ingredient_id and sub_menu_item_id are set or both null
    invalid_bom_entries = [
        # Both set
        {
            "name": "Invalid BOM Both Set",
            "price": 9.99,
            "category": "Test Category",
            "availability": True,
            "bom": [
                {
                    "ingredient_id": "ing123",
                    "sub_menu_item_id": "menu456",
                    "quantity": 1
                }
            ]
        },
        # Both null (no ingredient_id and no sub_menu_item_id)
        {
            "name": "Invalid BOM Both Null",
            "price": 9.99,
            "category": "Test Category",
            "availability": True,
            "bom": [
                {
                    # missing both fields
                    "quantity": 1
                }
            ]
        }
    ]

    for payload in invalid_bom_entries:
        response = None
        try:
            response = requests.post(
                BASE_URL + MENU_ENDPOINT,
                headers=headers,
                json=payload,
                timeout=30
            )
            assert response.status_code == 400, (
                f"Expected 400 Bad Request but got {response.status_code} "
                f"for payload: {payload}"
            )
        except requests.exceptions.RequestException as e:
            raise AssertionError(f"Request failed: {e}")