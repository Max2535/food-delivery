import requests
import uuid

BASE_URL = "http://localhost:8080"
REGISTER_PATH = "/v1/auth/register"
TIMEOUT = 30

def test_post_v1_auth_register_with_valid_data():
    url = BASE_URL + REGISTER_PATH
    # Generate unique user data to avoid conflicts
    unique_suffix = uuid.uuid4().hex[:8]
    payload = {
        "username": f"testuser_{unique_suffix}",
        "password": "StrongPassw0rd!",
        "email": f"testuser_{unique_suffix}@example.com"
    }
    headers = {
        "Content-Type": "application/json"
    }

    try:
        response = requests.post(url, json=payload, headers=headers, timeout=TIMEOUT)
        # Assert status code 201 Created
        assert response.status_code == 201, f"Expected status code 201, got {response.status_code}"

        json_resp = response.json()
        # Assert presence of user_id in response body
        assert "user_id" in json_resp, "Response JSON does not contain 'user_id'"

        user_id = json_resp["user_id"]
        assert isinstance(user_id, (int, str)), "'user_id' is not of expected type int or str"

    except requests.RequestException as e:
        assert False, f"Request failed: {str(e)}"

test_post_v1_auth_register_with_valid_data()
