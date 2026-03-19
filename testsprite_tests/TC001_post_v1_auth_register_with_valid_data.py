import requests
import uuid

BASE_URL = "http://localhost:8080"
REGISTER_ENDPOINT = "/v1/auth/register"
TIMEOUT = 30

def test_post_v1_auth_register_with_valid_data():
    url = BASE_URL + REGISTER_ENDPOINT
    # Generate unique username and email using uuid to avoid conflict
    unique_suffix = uuid.uuid4().hex[:8]
    payload = {
        "username": f"testuser_{unique_suffix}",
        "password": "ValidPass123!",
        "email": f"testuser_{unique_suffix}@example.com"
    }
    headers = {
        "Content-Type": "application/json"
    }
    try:
        response = requests.post(url, json=payload, headers=headers, timeout=TIMEOUT)
        assert response.status_code == 201, f"Expected status code 201, got {response.status_code}"
        response_json = response.json()
        assert "user_id" in response_json, "Response JSON does not contain user_id"
        assert isinstance(response_json["user_id"], (str, int)), "user_id should be string or integer"
    except requests.RequestException as e:
        raise AssertionError(f"Request failed: {e}")

test_post_v1_auth_register_with_valid_data()