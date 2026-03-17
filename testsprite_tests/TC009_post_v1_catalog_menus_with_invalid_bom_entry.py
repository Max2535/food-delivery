import requests

BASE_URL = "http://localhost:8080"
LOGIN_ENDPOINT = "/v1/auth/login"
MENU_ENDPOINT = "/v1/catalog/menus"
TIMEOUT = 30

# Use valid credentials for login to obtain JWT token
VALID_USERNAME = "testuser"
VALID_PASSWORD = "testpassword"


def test_post_v1_catalog_menus_with_invalid_bom_entry():
    # Step 1: Obtain JWT token
    login_payload = {
        "username": VALID_USERNAME,
        "password": VALID_PASSWORD
    }
    try:
        login_response = requests.post(
            f"{BASE_URL}{LOGIN_ENDPOINT}",
            json=login_payload,
            timeout=TIMEOUT
        )
        assert login_response.status_code == 200, f"Login failed with status {login_response.status_code}"
        token = login_response.json().get("token")
        assert token, "JWT token not found in login response"

        headers = {
            "Authorization": f"Bearer {token}",
            "Content-Type": "application/json"
        }

        # Step 2: Prepare invalid BOM entries to test both cases
        # Case A: Both ingredient_id and sub_menu_item_id are set (invalid)
        invalid_bom_both_set = [
            {
                "ingredient_id": "ingredient123",
                "sub_menu_item_id": "submenu456",
                "quantity": 2
            }
        ]

        # Case B: Both ingredient_id and sub_menu_item_id are null (invalid)
        invalid_bom_both_null = [
            {
                "quantity": 3
            }
        ]

        # Common minimal valid menu payload fields except BOM to isolate BOM errors
        base_menu_payload = {
            "name": "Test Menu Invalid BOM",
            "price": 9.99,
            "category": "Test Category",
            "availability": True
        }

        def post_menu_and_assert(bom):
            payload = dict(base_menu_payload)
            payload["bom"] = bom
            response = requests.post(
                f"{BASE_URL}{MENU_ENDPOINT}",
                json=payload,
                headers=headers,
                timeout=TIMEOUT
            )
            assert response.status_code == 400, f"Expected 400 Bad Request but got {response.status_code}"
            # Optionally check error message or response structure here if known

        # Test invalid BOM where both ingredient_id and sub_menu_item_id are set
        post_menu_and_assert(invalid_bom_both_set)

        # Test invalid BOM where both ingredient_id and sub_menu_item_id are null
        post_menu_and_assert(invalid_bom_both_null)

    except requests.RequestException as e:
        assert False, f"Request failed: {str(e)}"


test_post_v1_catalog_menus_with_invalid_bom_entry()