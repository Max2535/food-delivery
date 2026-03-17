import requests

def test_get_v1_kitchen_status_with_nonexistent_order_id():
    base_url = "http://localhost:8080"
    nonexistent_order_id = "nonexistent-order-12345"
    url = f"{base_url}/v1/kitchen/status/{nonexistent_order_id}"
    headers = {
        "Authorization": "Bearer your_token_here"
    }
    try:
        response = requests.get(url, headers=headers, timeout=30)
    except requests.RequestException as e:
        assert False, f"Request failed: {e}"
    assert response.status_code == 404, f"Expected 404 Not Found, got {response.status_code}"

test_get_v1_kitchen_status_with_nonexistent_order_id()