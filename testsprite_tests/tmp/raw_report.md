
# TestSprite AI Testing Report(MCP)

---

## 1️⃣ Document Metadata
- **Project Name:** food-delivery
- **Date:** 2026-03-17
- **Prepared by:** TestSprite AI Team

---

## 2️⃣ Requirement Validation Summary

#### Test TC001 post v1 auth register with valid data
- **Test Code:** [TC001_post_v1_auth_register_with_valid_data.py](./TC001_post_v1_auth_register_with_valid_data.py)
- **Test Error:** Traceback (most recent call last):
  File "/var/task/handler.py", line 258, in run_with_retry
    exec(code, exec_env)
  File "<string>", line 36, in <module>
  File "<string>", line 28, in test_post_v1_auth_register_with_valid_data
AssertionError: Response JSON does not contain 'user_id'

- **Test Visualization and Result:** https://www.testsprite.com/dashboard/mcp/tests/c9909424-ab57-4b08-bc40-19887e02053a/0441c38d-3051-4d4f-9e3c-04989dc4fd03
- **Status:** ❌ Failed
- **Analysis / Findings:** {{TODO:AI_ANALYSIS}}.
---

#### Test TC002 post v1 auth login with valid credentials
- **Test Code:** [TC002_post_v1_auth_login_with_valid_credentials.py](./TC002_post_v1_auth_login_with_valid_credentials.py)
- **Test Error:** Traceback (most recent call last):
  File "/var/task/handler.py", line 258, in run_with_retry
    exec(code, exec_env)
  File "<string>", line 35, in <module>
  File "<string>", line 24, in test_post_v1_auth_login_with_valid_credentials
AssertionError: Expected status code 200, got 401

- **Test Visualization and Result:** https://www.testsprite.com/dashboard/mcp/tests/c9909424-ab57-4b08-bc40-19887e02053a/6a5e0eb3-acdd-42b8-b5ce-35339b713a7c
- **Status:** ❌ Failed
- **Analysis / Findings:** {{TODO:AI_ANALYSIS}}.
---

#### Test TC003 post v1 auth login with invalid credentials
- **Test Code:** [TC003_post_v1_auth_login_with_invalid_credentials.py](./TC003_post_v1_auth_login_with_invalid_credentials.py)
- **Test Visualization and Result:** https://www.testsprite.com/dashboard/mcp/tests/c9909424-ab57-4b08-bc40-19887e02053a/dd5b4165-023f-4fe6-ad26-70522c9f1f07
- **Status:** ✅ Passed
- **Analysis / Findings:** {{TODO:AI_ANALYSIS}}.
---

#### Test TC004 post v1 orders with valid authorization and payload
- **Test Code:** [TC004_post_v1_orders_with_valid_authorization_and_payload.py](./TC004_post_v1_orders_with_valid_authorization_and_payload.py)
- **Test Error:** Traceback (most recent call last):
  File "/var/task/handler.py", line 258, in run_with_retry
    exec(code, exec_env)
  File "<string>", line 55, in <module>
  File "<string>", line 17, in test_post_v1_orders_with_valid_authorization_and_payload
AssertionError: Expected 200 OK on login, got 401

- **Test Visualization and Result:** https://www.testsprite.com/dashboard/mcp/tests/c9909424-ab57-4b08-bc40-19887e02053a/9bb27405-ad7d-4e6d-9423-9e7257423b4f
- **Status:** ❌ Failed
- **Analysis / Findings:** {{TODO:AI_ANALYSIS}}.
---

#### Test TC005 post v1 orders without authorization header
- **Test Code:** [TC005_post_v1_orders_without_authorization_header.py](./TC005_post_v1_orders_without_authorization_header.py)
- **Test Visualization and Result:** https://www.testsprite.com/dashboard/mcp/tests/c9909424-ab57-4b08-bc40-19887e02053a/972765da-697e-4125-b1a2-15cb62688f9a
- **Status:** ✅ Passed
- **Analysis / Findings:** {{TODO:AI_ANALYSIS}}.
---

#### Test TC006 get v1 kitchen status with valid order id
- **Test Code:** [TC006_get_v1_kitchen_status_with_valid_order_id.py](./TC006_get_v1_kitchen_status_with_valid_order_id.py)
- **Test Error:** Traceback (most recent call last):
  File "<string>", line 26, in test_get_v1_kitchen_status_with_valid_order_id
  File "/var/task/requests/models.py", line 1024, in raise_for_status
    raise HTTPError(http_error_msg, response=self)
requests.exceptions.HTTPError: 401 Client Error: Unauthorized for url: http://localhost:8080/v1/auth/login

During handling of the above exception, another exception occurred:

Traceback (most recent call last):
  File "/var/task/handler.py", line 258, in run_with_retry
    exec(code, exec_env)
  File "<string>", line 81, in <module>
  File "<string>", line 30, in test_get_v1_kitchen_status_with_valid_order_id
AssertionError: Failed to login to get token: 401 Client Error: Unauthorized for url: http://localhost:8080/v1/auth/login

- **Test Visualization and Result:** https://www.testsprite.com/dashboard/mcp/tests/c9909424-ab57-4b08-bc40-19887e02053a/c406c77f-017f-4f10-8c21-44059f1ced62
- **Status:** ❌ Failed
- **Analysis / Findings:** {{TODO:AI_ANALYSIS}}.
---

#### Test TC007 get v1 kitchen status with nonexistent order id
- **Test Code:** [TC007_get_v1_kitchen_status_with_nonexistent_order_id.py](./TC007_get_v1_kitchen_status_with_nonexistent_order_id.py)
- **Test Error:** Traceback (most recent call last):
  File "/var/task/handler.py", line 258, in run_with_retry
    exec(code, exec_env)
  File "<string>", line 22, in <module>
  File "<string>", line 17, in test_get_kitchen_status_with_nonexistent_order_id
AssertionError: Expected status code 404, got 400. Response body: {"error":"OrderID ไม่ถูกต้อง"}

- **Test Visualization and Result:** https://www.testsprite.com/dashboard/mcp/tests/c9909424-ab57-4b08-bc40-19887e02053a/91f765af-f4c1-4717-b00b-91ff05bb6084
- **Status:** ❌ Failed
- **Analysis / Findings:** {{TODO:AI_ANALYSIS}}.
---

#### Test TC008 post v1 catalog menus with valid authorization and payload
- **Test Code:** [TC008_post_v1_catalog_menus_with_valid_authorization_and_payload.py](./TC008_post_v1_catalog_menus_with_valid_authorization_and_payload.py)
- **Test Error:** Traceback (most recent call last):
  File "/var/task/handler.py", line 258, in run_with_retry
    exec(code, exec_env)
  File "<string>", line 92, in <module>
  File "<string>", line 46, in test_post_v1_catalog_menus_with_valid_authorization_and_payload
  File "<string>", line 38, in register_user
AssertionError: Registration response does not contain 'user_id'

- **Test Visualization and Result:** https://www.testsprite.com/dashboard/mcp/tests/c9909424-ab57-4b08-bc40-19887e02053a/309b398d-0642-4a23-9fbd-7d812388b77a
- **Status:** ❌ Failed
- **Analysis / Findings:** {{TODO:AI_ANALYSIS}}.
---

#### Test TC009 post v1 catalog menus with invalid bom entry
- **Test Code:** [TC009_post_v1_catalog_menus_with_invalid_bom_entry.py](./TC009_post_v1_catalog_menus_with_invalid_bom_entry.py)
- **Test Error:** Traceback (most recent call last):
  File "/var/task/handler.py", line 258, in run_with_retry
    exec(code, exec_env)
  File "<string>", line 81, in <module>
  File "<string>", line 25, in test_post_v1_catalog_menus_with_invalid_bom_entry
AssertionError: Login failed with status 401

- **Test Visualization and Result:** https://www.testsprite.com/dashboard/mcp/tests/c9909424-ab57-4b08-bc40-19887e02053a/376c49dd-6d80-4f33-9581-136463e9fdf9
- **Status:** ❌ Failed
- **Analysis / Findings:** {{TODO:AI_ANALYSIS}}.
---

#### Test TC010 post v1 inventory stock restock with valid authorization and payload
- **Test Code:** [TC010_post_v1_inventory_stock_restock_with_valid_authorization_and_payload.py](./TC010_post_v1_inventory_stock_restock_with_valid_authorization_and_payload.py)
- **Test Error:** Traceback (most recent call last):
  File "<string>", line 15, in get_jwt_token
  File "/var/task/requests/models.py", line 1024, in raise_for_status
    raise HTTPError(http_error_msg, response=self)
requests.exceptions.HTTPError: 401 Client Error: Unauthorized for url: http://localhost:8080/v1/auth/login

During handling of the above exception, another exception occurred:

Traceback (most recent call last):
  File "/var/task/handler.py", line 258, in run_with_retry
    exec(code, exec_env)
  File "<string>", line 64, in <module>
  File "<string>", line 41, in test_post_v1_inventory_stock_restock_with_valid_authorization_and_payload
  File "<string>", line 21, in get_jwt_token
RuntimeError: Failed to get JWT token: 401 Client Error: Unauthorized for url: http://localhost:8080/v1/auth/login

- **Test Visualization and Result:** https://www.testsprite.com/dashboard/mcp/tests/c9909424-ab57-4b08-bc40-19887e02053a/5866c489-a31f-4587-9060-c47c83e1e5b9
- **Status:** ❌ Failed
- **Analysis / Findings:** {{TODO:AI_ANALYSIS}}.
---


## 3️⃣ Coverage & Matching Metrics

- **20.00** of tests passed

| Requirement        | Total Tests | ✅ Passed | ❌ Failed  |
|--------------------|-------------|-----------|------------|
| ...                | ...         | ...       | ...        |
---


## 4️⃣ Key Gaps / Risks
{AI_GNERATED_KET_GAPS_AND_RISKS}
---