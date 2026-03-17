import requests

BASE_URL = "http://localhost:8080"
TIMEOUT = 30

# Replace this with a valid JWT token before running the test
JWT_TOKEN = "your_valid_jwt_token_here"


def test_post_v1_catalog_menus_with_invalid_bom_entry():
    url = f"{BASE_URL}/v1/catalog/menus"
    headers = {
        "Authorization": f"Bearer {JWT_TOKEN}",
        "Content-Type": "application/json"
    }

    # Prepare payload with invalid BOM entry: both ingredient_id and sub_menu_item_id are set
    invalid_bom_entry = {
        "name": "Invalid BOM Menu",
        "price": 10.99,
        "category": "Test Category",
        "availability": True,
        "bom": [
            {
                "ingredient_id": "some-ingredient-id",
                "sub_menu_item_id": "some-sub-menu-item-id",
                "quantity": 2
            }
        ]
    }

    try:
        response = requests.post(url, headers=headers, json=invalid_bom_entry, timeout=TIMEOUT)
        assert response.status_code == 400, f"Expected 400 Bad Request, got {response.status_code}"
    except requests.RequestException as e:
        assert False, f"Request to POST /v1/catalog/menus failed: {e}"


test_post_v1_catalog_menus_with_invalid_bom_entry()