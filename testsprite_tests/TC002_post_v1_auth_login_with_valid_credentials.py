import requests

BASE_URL = "http://localhost:8080"

def test_post_v1_auth_login_with_valid_credentials():
    url = f"{BASE_URL}/v1/auth/login"
    headers = {
        "Content-Type": "application/json"
    }
    # Use example valid credentials; adjust if needed for actual test environment
    payload = {
        "username": "validuser",
        "password": "validpassword"
    }
    try:
        response = requests.post(url, json=payload, headers=headers, timeout=30)
        response.raise_for_status()
    except requests.RequestException as e:
        assert False, f"Request failed: {e}"

    assert response.status_code == 200, f"Expected 200 OK but got {response.status_code}"
    
    json_resp = response.json()
    assert "token" in json_resp, "Response JSON does not contain 'token'"
    token = json_resp["token"]
    assert isinstance(token, str) and len(token) > 0, "JWT token should be a non-empty string"

test_post_v1_auth_login_with_valid_credentials()