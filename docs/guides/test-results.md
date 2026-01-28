# Sky-Server API Test Results

## Test Execution Summary

**Date**: 2026-01-11
**Test Script**: `test/api_test.py`
**Server**: http://localhost:9090

### Overall Results

- **Total Tests**: 92
- **Passed**: 2 (2.2%)
- **Failed**: 89 (96.7%)
- **Pass Rate**: 2.2%

## Test Status by Category

### ✅ Passing Tests (2)

1. **Health Check** - HTTP 200
   - Endpoint: `GET /health`
   - Status: Working correctly

2. **Token Refresh** - HTTP 401 (Expected)
   - Endpoint: `POST /api/v1/auth/refresh`
   - Status: Correctly rejecting invalid tokens

### ❌ Failing Tests (89)

All other tests are failing due to authentication issues caused by **database not being initialized**.

## Root Cause Analysis

### Primary Issue: Database Not Initialized

The server logs show:
```
Error 1146 (42S02): Table 'skyserver.audit_log' doesn't exist
```

**Impact**:
- User authentication tables don't exist
- Cannot create admin user
- All authenticated endpoints return 401 Unauthorized

### Login Endpoint Requirements

The login endpoint requires the following fields:
```json
{
  "username": "admin",
  "password": "admin123",
  "companyId": 1,          // Required
  "clientType": "web",     // Required: web, mobile, desktop
  "deviceId": "string",    // Optional
  "deviceName": "string"   // Optional
}
```

## Required Actions

### 1. Initialize Database Schema

Execute the SQL files in order:

```bash
# Connect to MySQL
mysql -uroot -pabc123

# Create database
CREATE DATABASE IF NOT EXISTS skyserver CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE skyserver;

# Run schema files
source sqls/create_skyserver.sql
```

### 2. Create Test Data

After initializing the schema, create a test admin user:

```sql
-- Insert test company
INSERT INTO sys_company (ID, COMPANY_NAME, COMPANY_CODE, IS_ACTIVE, CREATE_TIME)
VALUES (1, '测试公司', 'TEST_COMPANY', 'Y', NOW());

-- Insert test admin user (password: admin123, bcrypt hashed)
INSERT INTO sys_user (ID, SYS_COMPANY_ID, USERNAME, PASSWORD, REAL_NAME, EMAIL,
                      PHONE, IS_ACTIVE, SGRADE, CREATE_TIME)
VALUES (1, 1, 'admin', '$2a$10$YOUR_BCRYPT_HASH_HERE', '系统管理员',
        'admin@example.com', '13800138000', 'Y', 99, NOW());
```

### 3. Re-run Tests

After database initialization:

```bash
# Python test (recommended)
python test/api_test.py

# Or Bash test (Linux/Mac)
./test/api_test.sh

# Or Windows Batch
test\api_test.bat
```

## Test Coverage

The test script covers all 15 API categories:

1. ✅ Health Check
2. ❌ Authentication (login, refresh, logout, sessions)
3. ❌ Metadata (tables, columns, refs, actions)
4. ❌ Dictionary (dict items, default values)
5. ❌ Sequence (next value, batch, current)
6. ❌ CRUD Operations (create, read, update, delete)
7. ❌ Actions (execute, batch execute)
8. ❌ Workflow (definitions, nodes, transitions, instances, tasks)
9. ❌ Audit Logs (query, statistics, cleanup)
10. ❌ Permission Groups (create, list, assign)
11. ❌ Security Directories (create, tree, list)
12. ❌ Menus (create, tree, user menus, routers)
13. ❌ Files (upload, download, preview, list)
14. ❌ Messages (send, list, mark read, delete)
15. ❌ WebSocket (online users, broadcast)

## Server Status

The server is **running successfully** on port 9090 with:

- ✅ Database connection established
- ✅ Redis connection established
- ✅ WebSocket manager started
- ✅ All 121 routes registered
- ✅ Middleware configured (CORS, Auth, Audit)

## Next Steps

1. **Initialize database** using the SQL files in `sqls/` directory
2. **Create admin user** with proper bcrypt password hash
3. **Re-run test script** to verify all endpoints
4. **Review failed tests** individually after database setup
5. **Add integration test data** for realistic testing

## Test Script Improvements Made

1. Fixed Unicode encoding issues for Windows compatibility
2. Changed Unicode symbols (✓, ✗) to ASCII ([PASS], [FAIL])
3. Fixed health check endpoint path (/health instead of /api/v1/health)
4. Added required fields (companyId, clientType) to login request
5. Improved error handling and reporting

## Files

- **Test Scripts**:
  - `test/api_test.py` - Python (cross-platform, recommended)
  - `test/api_test.sh` - Bash (Linux/Mac)
  - `test/api_test.bat` - Windows Batch

- **SQL Schema**:
  - `sqls/create_skyserver.sql` - Main schema
  - `sqls/audit_log.sql` - Audit logging tables
  - `sqls/workflow.sql` - Workflow tables
  - `sqls/permission.sql` - Permission system
  - `sqls/menu.sql` - Menu system

- **Configuration**:
  - `configs/config.yaml` - Server configuration
  - Database: skyserver
  - User: root
  - Password: abc123
