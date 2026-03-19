import requests

def test_post_v1_auth_login_with_invalid_credentials():
    base_url = "http://localhost:8080"
    endpoint = "/v1/auth/login"
    url = base_url + endpoint

    # Use invalid username and password
    payload = {
        "username": "invalid_user_12345",
        "password": "wrong_password_67890"
    }
    headers = {
        "Content-Type": "application/json",
        "authType": "Bearer token"
    }

    try:
        response = requests.post(url, json=payload, headers=headers, timeout=30)
    except requests.RequestException as e:
        assert False, f"Request failed: {e}"

    # Assert that response status code is 401 Unauthorized
    assert response.status_code == 401, (
        f"Expected status code 401 Unauthorized, got {response.status_code}. "
        f"Response body: {response.text}"
    )

test_post_v1_auth_login_with_invalid_credentials()