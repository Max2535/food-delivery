import requests

BASE_URL = "http://localhost:8080"
NONEXISTENT_ORDER_ID = "123e4567-e89b-12d3-a456-426614174000"  # valid UUID format
TIMEOUT = 30

def test_get_kitchen_status_with_nonexistent_order_id():
    url = f"{BASE_URL}/v1/kitchen/status/{NONEXISTENT_ORDER_ID}"
    headers = {
        "Accept": "application/json"
    }
    try:
        response = requests.get(url, headers=headers, timeout=TIMEOUT)
    except requests.RequestException as e:
        assert False, f"Request failed: {e}"

    assert response.status_code == 404, (
        f"Expected status code 404, got {response.status_code}. "
        f"Response body: {response.text}"
    )

test_get_kitchen_status_with_nonexistent_order_id()
