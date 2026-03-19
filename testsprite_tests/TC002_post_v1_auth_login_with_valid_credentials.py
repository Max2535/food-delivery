import requests

BASE_URL = "http://localhost:8080"
LOGIN_ENDPOINT = "/v1/auth/login"
TIMEOUT = 30

def test_post_v1_auth_login_with_valid_credentials():
    url = BASE_URL + LOGIN_ENDPOINT
    headers = {
        "Content-Type": "application/json"
    }
    # Replace with valid test credentials
    payload = {
        "username": "validuser",
        "password": "validpassword"
    }
    try:
        response = requests.post(url, json=payload, headers=headers, timeout=TIMEOUT)
        assert response.status_code == 200, f"Expected status code 200 but got {response.status_code}"
        json_resp = response.json()
        # Validate that JWT token is present in response body (e.g. a token field)
        assert "token" in json_resp or "jwt" in json_resp or "access_token" in json_resp, "JWT token not found in response"
        # Check token is non-empty string
        token = json_resp.get("token") or json_resp.get("jwt") or json_resp.get("access_token")
        assert isinstance(token, str) and token.strip() != "", "JWT token is empty or invalid"
    except requests.RequestException as e:
        assert False, f"RequestException occurred: {e}"

test_post_v1_auth_login_with_valid_credentials()
