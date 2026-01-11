-- Sky-Server Database Initialization Script
-- This script creates the database and initializes all schemas

-- Create database
CREATE DATABASE IF NOT EXISTS skyserver CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE skyserver;

-- Display database info
SELECT 'Database skyserver created/selected' AS Status;

-- Source main schema
SOURCE create_skyserver.sql;
SELECT 'Main schema loaded' AS Status;

-- Source audit log schema
SOURCE audit_log.sql;
SELECT 'Audit log schema loaded' AS Status;

-- Source workflow schema
SOURCE workflow.sql;
SELECT 'Workflow schema loaded' AS Status;

-- Source permission schema
SOURCE permission.sql;
SELECT 'Permission schema loaded' AS Status;

-- Source menu schema
SOURCE menu.sql;
SELECT 'Menu schema loaded' AS Status;

-- Create test company
INSERT INTO sys_company (ID, COMPANY_NAME, COMPANY_CODE, CONTACT_PERSON, CONTACT_PHONE,
                         ADDRESS, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME)
VALUES (1, '测试公司', 'TEST_COMPANY', '管理员', '13800138000',
        '测试地址', 'Y', 1, NOW(), 1, NOW())
ON DUPLICATE KEY UPDATE COMPANY_NAME = '测试公司';

SELECT 'Test company created' AS Status;

-- Create test admin user
-- Password: admin123 (bcrypt hash)
-- Note: This is a bcrypt hash of "admin123" with cost 10
INSERT INTO sys_user (ID, SYS_COMPANY_ID, USERNAME, PASSWORD, REAL_NAME, EMAIL, PHONE,
                      IS_ACTIVE, SGRADE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME)
VALUES (1, 1, 'admin',
        '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
        '系统管理员', 'admin@example.com', '13800138000',
        'Y', 99, 1, NOW(), 1, NOW())
ON DUPLICATE KEY UPDATE PASSWORD = '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy';

SELECT 'Admin user created (username: admin, password: admin123)' AS Status;

-- Display summary
SELECT
    'Database Initialization Complete!' AS Status,
    COUNT(*) as TableCount
FROM information_schema.tables
WHERE table_schema = 'skyserver';

SELECT 'Ready for testing!' AS Status;



-- Insert test data for Sky-Server
USE skyserver;

-- Create test company
INSERT INTO sys_company (ID, SYS_COMPANY_ID, NAME, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME)
VALUES (1, 1, '测试公司', 'Y', '1', NOW(), '1', NOW())
    ON DUPLICATE KEY UPDATE NAME = '测试公司';

SELECT 'Test company created/updated' AS Status;

-- Create test admin user
-- Password: admin123 (bcrypt hash with cost 10)
-- $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy
INSERT INTO sys_user (ID, SYS_COMPANY_ID, USERNAME, PASSWORD, TRUE_NAME, EMAIL, PHONE,
                      IS_ACTIVE, IS_ADMIN, SGRADE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME)
VALUES (1, 1, 'admin',
        '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
        '系统管理员', 'admin@example.com', '13800138000',
        'Y', 'Y', 99, '1', NOW(), '1', NOW())
    ON DUPLICATE KEY UPDATE
                         PASSWORD = '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
                         TRUE_NAME = '系统管理员',
                         IS_ADMIN = 'Y',
                         SGRADE = 99;

SELECT 'Admin user created/updated (username: admin, password: admin123)' AS Status;

-- Verify data
SELECT
    'Database Ready!' AS Status,
    (SELECT COUNT(*) FROM sys_company) AS Companies,
    (SELECT COUNT(*) FROM sys_user) AS Users;



-- Insert test data for Sky-Server (ASCII only to avoid encoding issues)
USE skyserver;

-- Set charset
SET NAMES utf8mb4;
SET CHARACTER SET utf8mb4;

-- Create test company
INSERT INTO sys_company (ID, SYS_COMPANY_ID, NAME, IS_ACTIVE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME)
VALUES (1, 1, 'Test Company', 'Y', '1', NOW(), '1', NOW())
    ON DUPLICATE KEY UPDATE NAME = 'Test Company';

SELECT 'Test company created/updated' AS Status;

-- Create test admin user
-- Password: admin123 (bcrypt hash with cost 10)
-- Hash generated from: admin123
INSERT INTO sys_user (ID, SYS_COMPANY_ID, USERNAME, PASSWORD, TRUE_NAME, EMAIL, PHONE,
                      IS_ACTIVE, IS_ADMIN, SGRADE, CREATE_BY, CREATE_TIME, UPDATE_BY, UPDATE_TIME)
VALUES (1, 1, 'admin',
        '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
        'System Administrator', 'admin@example.com', '13800138000',
        'Y', 'Y', 99, '1', NOW(), '1', NOW())
    ON DUPLICATE KEY UPDATE
                         PASSWORD = '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy',
                         TRUE_NAME = 'System Administrator',
                         IS_ADMIN = 'Y',
                         SGRADE = 99;

SELECT 'Admin user created/updated (username: admin, password: admin123)' AS Status;

-- Verify data
SELECT
    'Database Ready!' AS Status,
    (SELECT COUNT(*) FROM sys_company) AS Companies,
    (SELECT COUNT(*) FROM sys_user) AS Users;


-- Update admin password with correct bcrypt hash
USE skyserver;

UPDATE sys_user
SET PASSWORD = '$2a$10$iztoR7MeHZKyoBNpJM4pjOZ729KAoy.5x5ayetl1Rnb3TBgVCO0jy'
WHERE USERNAME = 'admin';

SELECT 'Password updated successfully' AS Status;

-- Verify update
SELECT ID, USERNAME, LEFT(PASSWORD, 30) AS Password_Prefix, TRUE_NAME
FROM sys_user
WHERE USERNAME = 'admin';
