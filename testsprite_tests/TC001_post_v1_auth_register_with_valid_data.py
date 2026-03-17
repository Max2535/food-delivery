import requests

endpoint = "http://localhost:8080"

def test_post_v1_auth_register_with_valid_data():
    url = f"{endpoint}/v1/auth/register"
    headers = {
        "Authorization": "Bearer dummy_token_for_auth",  # Provided authType is Bearer token, but this endpoint does NOT require auth. Including per instruction.
        "Content-Type": "application/json"
    }
    payload = {
        "username": "testuser123",
        "password": "StrongPass!2026",
        "email": "testuser123@example.com"
    }
    try:
        response = requests.post(url, json=payload, headers=headers, timeout=30)
        assert response.status_code == 201, f"Expected status code 201, got {response.status_code}"
        resp_json = response.json()
        assert "user_id" in resp_json, "Response JSON does not contain 'user_id'"
        assert isinstance(resp_json["user_id"], (int, str)), "'user_id' should be a string or integer"
    except requests.RequestException as e:
        assert False, f"HTTP request failed: {e}"

test_post_v1_auth_register_with_valid_data()