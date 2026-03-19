import requests

BASE_URL = "http://localhost:8080"
# Please replace the token below with a valid JWT token for authentication
AUTH_TOKEN = "Bearer REPLACE_WITH_VALID_JWT_TOKEN"

def test_post_v1_catalog_menus_with_invalid_bom_entry():
    url = f"{BASE_URL}/v1/catalog/menus"
    headers = {
        "Authorization": AUTH_TOKEN,
        "Content-Type": "application/json"
    }
    
    # Define invalid BOM payloads where ingredient_id and sub_menu_item_id are both set or both null.
    invalid_bom_entries = [
        {
            # Both ingredient_id and sub_menu_item_id set
            "name": "Test Menu Invalid BOM Both Set",
            "price": 9.99,
            "category": "Test Category",
            "availability": True,
            "bom": [
                {
                    "ingredient_id": "ing-123",
                    "sub_menu_item_id": "sub-456",
                    "quantity": 2
                }
            ]
        },
        {
            # Both ingredient_id and sub_menu_item_id null
            "name": "Test Menu Invalid BOM Both Null",
            "price": 12.99,
            "category": "Test Category",
            "availability": True,
            "bom": [
                {
                    "ingredient_id": None,
                    "sub_menu_item_id": None,
                    "quantity": 1
                }
            ]
        }
    ]

    for payload in invalid_bom_entries:
        try:
            response = requests.post(url, headers=headers, json=payload, timeout=30)
        except requests.RequestException as e:
            assert False, f"Request failed with exception: {e}"
        if response.status_code == 401:
            assert False, "Unauthorized: Please provide a valid JWT token in AUTH_TOKEN to run this test."
        # Expecting HTTP 400 Bad Request for invalid BOM entries
        assert response.status_code == 400, f"Expected 400 Bad Request but got {response.status_code}. Response: {response.text}"

test_post_v1_catalog_menus_with_invalid_bom_entry()
