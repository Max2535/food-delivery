import requests

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

def test_get_v1_kitchen_status_with_nonexistent_order_id():
    nonexistent_order_id = "nonexistent_order_12345"
    url = f"{BASE_URL}/v1/kitchen/status/{nonexistent_order_id}"
    try:
        response = requests.get(url, timeout=TIMEOUT)
    except requests.RequestException as e:
        assert False, f"Request failed: {e}"
    assert response.status_code == 404, f"Expected 404 Not Found, got {response.status_code}"
    
test_get_v1_kitchen_status_with_nonexistent_order_id()