import requests

BASE_URL = "http://localhost:8080"
LOGIN_ENDPOINT = "/v1/auth/login"
TIMEOUT = 30

def test_post_v1_auth_login_with_valid_credentials():
    url = BASE_URL + LOGIN_ENDPOINT

    # Example valid credentials - update these as needed
    payload = {
        "username": "validUser",
        "password": "validPassword123!"
    }
    headers = {
        "Content-Type": "application/json"
    }

    try:
        response = requests.post(url, json=payload, headers=headers, timeout=TIMEOUT)
    except requests.RequestException as e:
        assert False, f"Request failed: {e}"

    assert response.status_code == 200, f"Expected status code 200, got {response.status_code}"

    try:
        data = response.json()
    except ValueError:
        assert False, "Response is not valid JSON"

    assert "token" in data, "JWT token not found in response"
    token = data["token"]
    assert isinstance(token, str) and len(token) > 0, "JWT token is empty or not a string"

test_post_v1_auth_login_with_valid_credentials()