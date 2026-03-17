import requests

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

def test_post_v1_auth_login_with_invalid_credentials():
    url = f"{BASE_URL}/v1/auth/login"
    headers = {
        "Content-Type": "application/json"
    }
    payload = {
        "username": "invalid_user",
        "password": "wrong_password"
    }
    try:
        response = requests.post(url, json=payload, headers=headers, timeout=TIMEOUT)
    except requests.RequestException as e:
        assert False, f"Request failed: {e}"

    assert response.status_code == 401, f"Expected 401 Unauthorized, got {response.status_code}"

test_post_v1_auth_login_with_invalid_credentials()