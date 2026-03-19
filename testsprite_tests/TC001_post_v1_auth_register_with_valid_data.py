import requests
import uuid

BASE_URL = "http://localhost:8080"
REGISTER_ENDPOINT = "/v1/auth/register"
TIMEOUT = 30

def test_post_v1_auth_register_with_valid_data():
    url = BASE_URL + REGISTER_ENDPOINT
    # Generate unique username and email for test isolation
    unique_suffix = str(uuid.uuid4()).split('-')[0]
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
        assert response.status_code == 201, f"Expected 201 Created, got {response.status_code}, response: {response.text}"
        json_response = response.json()
        assert "user_id" in json_response, f"'user_id' not found in response body: {json_response}"
        assert isinstance(json_response["user_id"], (str, int)), f"user_id type is not str or int: {type(json_response['user_id'])}"
    except requests.RequestException as e:
        assert False, f"HTTP Request failed: {e}"

test_post_v1_auth_register_with_valid_data()