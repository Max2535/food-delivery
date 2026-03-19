import requests

BASE_URL = "http://localhost:8080"

def test_post_v1_orders_without_authorization_header():
    url = f"{BASE_URL}/v1/orders"
    headers = {
        "Content-Type": "application/json"
    }
    payload = {
        "customer_id": "test_customer_id",
        "items": [
            {
                "menu_item_id": "test_menu_item_id",
                "quantity": 1
            }
        ],
        "total_amount": 9.99
    }

    try:
        response = requests.post(url, json=payload, headers=headers, timeout=30)
    except requests.RequestException as e:
        raise AssertionError(f"Request failed: {e}")

    assert response.status_code == 401, (
        f"Expected status code 401 Unauthorized but got {response.status_code}. "
        f"Response body: {response.text}"
    )

test_post_v1_orders_without_authorization_header()