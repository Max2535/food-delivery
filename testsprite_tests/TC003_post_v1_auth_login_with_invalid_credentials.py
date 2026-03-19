import requests

BASE_URL = "http://localhost:8080"

def test_post_v1_auth_login_with_invalid_credentials():
    url = f"{BASE_URL}/v1/auth/login"
    headers = {
        "Content-Type": "application/json"
    }
    # Using invalid username and password
    payload = {
        "username": "invalid_user_123",
        "password": "wrong_password_456"
    }
    try:
        response = requests.post(url, json=payload, headers=headers, timeout=30)
        assert response.status_code == 401, f"Expected 401 Unauthorized but got {response.status_code}. Response body: {response.text}"
    except requests.RequestException as e:
        assert False, f"Request to {url} failed with exception: {e}"

test_post_v1_auth_login_with_invalid_credentials()