import requests

BASE_URL = "http://localhost:8080"
AUTH_REGISTER_PATH = "/v1/auth/register"
AUTH_LOGIN_PATH = "/v1/auth/login"
CREATE_MENU_PATH = "/v1/catalog/menus"
TIMEOUT = 30

# Test credentials used for registration and login
TEST_USERNAME = "testuser"
TEST_PASSWORD = "testpassword"
TEST_EMAIL = "testuser@example.com"

def test_post_v1_catalog_menus_with_valid_authorization_and_payload():
    # Register a user to ensure login succeeds
    register_url = BASE_URL + AUTH_REGISTER_PATH
    register_payload = {
        "username": TEST_USERNAME,
        "password": TEST_PASSWORD,
        "email": TEST_EMAIL
    }
    try:
        register_response = requests.post(register_url, json=register_payload, timeout=TIMEOUT)
        # Accept 201 Created, 400 Bad Request, or 409 Conflict if user already exists
        assert register_response.status_code in (201, 400, 409), f"Registration failed with status {register_response.status_code}"
    except Exception as e:
        raise AssertionError(f"Registration request failed: {e}")

    # Login to get JWT token
    login_url = BASE_URL + AUTH_LOGIN_PATH
    login_payload = {
        "username": TEST_USERNAME,
        "password": TEST_PASSWORD
    }
    try:
        login_response = requests.post(login_url, json=login_payload, timeout=TIMEOUT)
        assert login_response.status_code == 200, f"Login failed with status {login_response.status_code}"
        login_data = login_response.json()
        token = login_data.get("token") or login_data.get("access_token")
        assert token and isinstance(token, str), "JWT token not found in login response"
    except Exception as e:
        raise AssertionError(f"Login request failed: {e}")

    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }

    # Define a valid menu payload
    menu_payload = {
        "name": "Test Menu Item",
        "description": "A test menu item description",
        "price": 9.99,
        "category": "Test Category",
        "availability": True
    }

    create_menu_url = BASE_URL + CREATE_MENU_PATH

    created_menu_id = None
    try:
        response = requests.post(create_menu_url, headers=headers, json=menu_payload, timeout=TIMEOUT)
        assert response.status_code == 201, f"Expected 201 Created but got {response.status_code}"
        resp_json = response.json()
        created_menu_id = resp_json.get("menu_id")
        assert created_menu_id is not None, "menu_id not found in create menu response"
    finally:
        # Clean up created menu if created_menu_id present
        if created_menu_id:
            try:
                del_response = requests.delete(f"{create_menu_url}/{created_menu_id}", headers=headers, timeout=TIMEOUT)
                # Accept 200 OK or 204 No Content on delete success
                assert del_response.status_code in (200, 204), f"Cleanup delete failed with status {del_response.status_code}"
            except Exception:
                pass


test_post_v1_catalog_menus_with_valid_authorization_and_payload()
