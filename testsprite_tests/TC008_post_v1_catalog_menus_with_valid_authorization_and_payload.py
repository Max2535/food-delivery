import requests

endpoint = "http://localhost:8080"
timeout = 30

def test_post_v1_catalog_menus_with_valid_authorization_and_payload():
    # This token should be replaced with a valid JWT token for testing
    jwt_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.ValidTestTokenExample"

    headers = {
        "Authorization": f"Bearer {jwt_token}",
        "Content-Type": "application/json"
    }

    menu_payload = {
        "name": "Test Menu Item",
        "price": 9.99,
        "category": "Main Course",
        "availability": True,
        "description": "A delicious test menu item",
        "bom": [  # Bill of Materials example - ingredient or sub_menu_item references
            {
                "ingredient_id": None,
                "sub_menu_item_id": None,
                "quantity": 1
            }
        ]
    }

    # Because the PRD does not specify full schema, adjust payload elements to plausible keys
    # Remove bom entry to valid minimal menu payload that should be accepted (assumed optional)
    # Or fix bom entry to valid one (ingredient_id or sub_menu_item_id must be set)
    # We'll fix bom entry by setting only ingredient_id to an example UUID string as string or number
    menu_payload["bom"] = [
        {
            "ingredient_id": "123e4567-e89b-12d3-a456-426614174000",
            "sub_menu_item_id": None,
            "quantity": 2
        }
    ]

    created_menu_id = None
    try:
        response = requests.post(
            f"{endpoint}/v1/catalog/menus",
            headers=headers,
            json=menu_payload,
            timeout=timeout
        )
        assert response.status_code == 201, f"Expected 201, got {response.status_code}: {response.text}"
        resp_json = response.json()
        assert "menu_id" in resp_json, "Response JSON does not contain 'menu_id'"
        created_menu_id = resp_json["menu_id"]
    finally:
        if created_menu_id:
            try:
                del_response = requests.delete(
                    f"{endpoint}/v1/catalog/menus/{created_menu_id}",
                    headers=headers,
                    timeout=timeout
                )
                # It's OK if delete fails here, no assertion raised
            except Exception:
                pass

test_post_v1_catalog_menus_with_valid_authorization_and_payload()