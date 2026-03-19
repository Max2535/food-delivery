import requests

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

def test_post_v1_auth_login_with_valid_credentials():
    url = f"{BASE_URL}/v1/auth/login"
    headers = {
        "Content-Type": "application/json"
    }
    # Use valid credentials - example credentials, should be replaced with real valid ones as per environment 
    payload = {
        "username": "validuser",
        "password": "validpassword"
    }
    try:
        response = requests.post(url, json=payload, headers=headers, timeout=TIMEOUT)
    except requests.RequestException as e:
        assert False, f"HTTP request failed: {e}"

    assert response.status_code == 200, f"Expected status code 200, got {response.status_code}"
    
    try:
        response_body = response.json()
    except ValueError:
        assert False, "Response is not valid JSON"

    assert "token" in response_body, "Response JSON does not contain 'token'"
    token = response_body["token"]
    assert isinstance(token, str) and len(token) > 0, "JWT token should be a non-empty string"

test_post_v1_auth_login_with_valid_credentials()