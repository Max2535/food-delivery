import requests

def test_get_v1_kitchen_status_with_nonexistent_order_id():
    base_url = "http://localhost:8080"
    nonexistent_order_id = "00000000-0000-0000-0000-000000000000"  # UUID unlikely to exist
    url = f"{base_url}/v1/kitchen/status/{nonexistent_order_id}"
    headers = {
        "Accept": "application/json"
    }
    try:
        response = requests.get(url, headers=headers, timeout=30)
        assert response.status_code == 404, f"Expected status code 404, got {response.status_code}"
    except requests.RequestException as e:
        assert False, f"Request failed: {e}"

test_get_v1_kitchen_status_with_nonexistent_order_id()