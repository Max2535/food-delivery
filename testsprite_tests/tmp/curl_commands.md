# Food Delivery API - CURL Commands

This document contains CURL commands for the test cases defined in the TestSprite report. These commands can be used for manual verification of the API endpoints.

> [!TIP]
> You can set the base URL and tokens as environment variables in your terminal for easier use:
> ```bash
> export BASE_URL="http://localhost:8080"
> export TOKEN="your_jwt_token_here"
> ```

---

## 1. Authentication

### TC001: Register with Valid Data
Registers a new user with a unique username and email.
```bash
curl -X POST "${BASE_URL:-http://localhost:8080}/v1/auth/register" \
     -H "Content-Type: application/json" \
     -d '{
           "username": "testuser_$(date +%s)",
           "password": "ValidPass123!",
           "email": "testuser_$(date +%s)@example.com"
         }'
```

### TC002: Login with Valid Credentials
Authenticates a user and returns a JWT token.
```bash
curl -X POST "${BASE_URL:-http://localhost:8080}/v1/auth/login" \
     -H "Content-Type: application/json" \
     -d '{
           "username": "validuser",
           "password": "validpassword"
         }'
```

### TC003: Login with Invalid Credentials
Attempts to login with incorrect credentials.
```bash
curl -X POST "${BASE_URL:-http://localhost:8080}/v1/auth/login" \
     -H "Content-Type: application/json" \
     -H "authType: Bearer token" \
     -d '{
           "username": "invalid_user_12345",
           "password": "wrong_password_67890"
         }'
```

---

## 2. Orders

### TC004: Create Order (Authorized)
Creates a new order for a customer. Requires a valid JWT token.
```bash
curl -X POST "${BASE_URL:-http://localhost:8080}/v1/orders" \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer ${TOKEN}" \
     -d '{
           "customer_id": "customer-123",
           "items": [
             {
               "menu_item_id": "menuitem-456",
               "quantity": 2
             }
           ],
           "total_amount": 25.50
         }'
```

### TC005: Create Order (Unauthorized)
Attempts to create an order without the Authorization header.
```bash
curl -X POST "${BASE_URL:-http://localhost:8080}/v1/orders" \
     -H "Content-Type: application/json" \
     -d '{
           "customer_id": "test-customer-id",
           "items": [
             {
               "menu_item_id": "test-menu-item-id",
               "quantity": 1,
               "portion_multiplier": 1.0
             }
           ],
           "total_amount": 15.50
         }'
```

---

## 3. Kitchen

### TC006: Get Kitchen Status (Valid Order ID)
Retrieves the status of a specific order in the kitchen.
```bash
# Replace <ORDER_ID> with a real ID from TC004 response
curl -X GET "${BASE_URL:-http://localhost:8080}/v1/kitchen/status/<ORDER_ID>"
```

### TC007: Get Kitchen Status (Nonexistent Order ID)
Attempts to retrieve status for an ID that does not exist.
```bash
curl -X GET "${BASE_URL:-http://localhost:8080}/v1/kitchen/status/nonexistent_order_12345"
```

---

## 4. Catalog

### TC008: Create Menu Item (Authorized)
Adds a new item to the menu catalog. Requires Authorization.
```bash
curl -X POST "${BASE_URL:-http://localhost:8080}/v1/catalog/menus" \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer ${TOKEN}" \
     -d '{
           "name": "Test Menu Item $(date +%s)",
           "price": 9.99,
           "category": "Test Category",
           "availability": true,
           "bom": [],
           "description": "Automated test menu item"
         }'
```

### TC009: Create Menu Item (Invalid BOM)
Attempts to create a menu item with invalid Bill of Materials (BOM) entries.
```bash
# Example 1: Both ingredient_id and sub_menu_item_id set
curl -X POST "${BASE_URL:-http://localhost:8080}/v1/catalog/menus" \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer ${TOKEN}" \
     -d '{
           "name": "Invalid BOM Both Set",
           "price": 9.99,
           "bom": [{"ingredient_id": "ing-123", "sub_menu_item_id": "sub-456", "quantity": 2}]
         }'
```

---

## 5. Inventory

### TC010: Restock Inventory
Restocks a specific material in the inventory.
```bash
# First, list materials to get a valid <MATERIAL_ID>
# curl -X GET "${BASE_URL:-http://localhost:8080}/v1/inventory/materials"

curl -X POST "${BASE_URL:-http://localhost:8080}/v1/inventory/stock/restock" \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer ${TOKEN}" \
     -d '{
           "material_id": "<MATERIAL_ID>",
           "quantity": 10,
           "note": "Test restock via CURL"
         }'
```
