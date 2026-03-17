import requests
import uuid

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

# Credentials for test user
test_username = f"testuser_{uuid.uuid4()}"
test_password = "StrongPass!123"
test_email = f"{test_username}@example.com"


def get_jwt_token(username, password):
    login_url = f"{BASE_URL}/v1/auth/login"
    login_payload = {
        "username": username,
        "password": password
    }
    resp = requests.post(login_url, json=login_payload, timeout=TIMEOUT)
    assert resp.status_code == 200, f"Login failed with status {resp.status_code} and response {resp.text}"
    data = resp.json()
    assert "token" in data, "Login response does not contain 'token'"
    token = data["token"]
    assert isinstance(token, str) and token.strip() != "", "JWT token is empty or not a string"
    return token


def register_user(username, password, email):
    register_url = f"{BASE_URL}/v1/auth/register"
    register_payload = {
        "username": username,
        "password": password,
        "email": email
    }
    resp = requests.post(register_url, json=register_payload, timeout=TIMEOUT)
    assert resp.status_code == 201, f"User registration failed with status {resp.status_code} and response {resp.text}"
    data = resp.json()
    assert "user_id" in data, "Registration response does not contain 'user_id'"
    user_id = data["user_id"]
    assert isinstance(user_id, str) and user_id.strip() != "", "user_id is empty or not a string"
    return user_id


def test_post_v1_catalog_menus_with_valid_authorization_and_payload():
    # Register a new user
    register_user(test_username, test_password, test_email)

    # Login to get a valid JWT token
    token = get_jwt_token(test_username, test_password)

    url = f"{BASE_URL}/v1/catalog/menus"
    headers = {
        "Authorization": f"Bearer {token}",
        "Content-Type": "application/json"
    }
    unique_name = f"Test Menu {uuid.uuid4()}"
    payload = {
        "name": unique_name,
        "description": "A test menu item created during automated testing.",
        "price": 12.99,
        "category": "Test Category",
        "availability": True,
        "bom": [
            {
                "ingredient_id": "ingredient-uuid-example-1234",
                "quantity": 1.0,
                "unit": "pcs"
            }
        ]
    }

    menu_id = None
    try:
        response = requests.post(url, headers=headers, json=payload, timeout=TIMEOUT)
        assert response.status_code == 201, f"Expected 201, got {response.status_code}, response: {response.text}"
        resp_json = response.json()
        assert "menu_id" in resp_json, "Response JSON does not contain 'menu_id'"
        menu_id = resp_json["menu_id"]
        assert isinstance(menu_id, str) and menu_id.strip() != "", "'menu_id' is empty or not a string"
    finally:
        if menu_id:
            delete_url = f"{BASE_URL}/v1/catalog/menus/{menu_id}"
            try:
                delete_headers = {
                    "Authorization": f"Bearer {token}"
                }
                requests.delete(delete_url, headers=delete_headers, timeout=TIMEOUT)
            except Exception:
                pass


test_post_v1_catalog_menus_with_valid_authorization_and_payload()
