# Database Schema

This directory contains the complete database initialization script for Sky-Server.

## Quick Start

### Initialize Database (One Command)

```bash
mysql -u root -p < sqls/init.sql
```

This single command will:
1. Create the `skyserver` database
2. Create all tables with proper schemas
3. Insert test company (ID=1)
4. Create admin user with default credentials

### Default Credentials

After initialization, login with:
- **Username**: `admin`
- **Password**: `admin123`
- **Company ID**: `1`

⚠️ **Security Warning**: Change the default password immediately in production!

## What's Included

The `init.sql` file contains:
- 20+ system tables (sys_*)
- Complete indexes and constraints
- Initial test data for development

### Key Tables
- `sys_company` - Multi-tenant company management
- `sys_user` - User accounts and authentication
- `sys_table` - Metadata table definitions
- `sys_column` - Metadata field definitions
- `sys_groups` - Permission groups
- `sys_directory` - Security directories
- `sys_action` - Custom actions
- `sys_table_cmd` - Table hooks/commands
- And more...

## Common Operations

### Reset Database
```bash
mysql -u root -p -e "DROP DATABASE IF EXISTS skyserver;"
mysql -u root -p < sqls/init.sql
```

### Backup Database
```bash
mysqldump -u root -p skyserver > backup_$(date +%Y%m%d).sql
```

### Check Tables
```bash
mysql -u root -p skyserver -e "SHOW TABLES;"
```

## Schema Migrations

For existing databases, use migration scripts when schema changes:

### 1. Rename sys_column.NAME to DISPLAY_NAME (2026-01-12)
```bash
mysql -u root -p skyserver < sqls/migration_name_to_display_name.sql
```

⚠️ **Required** if you have an existing database created before 2026-01-12.

This migration renames the `NAME` column to `DISPLAY_NAME` in the `sys_column` table for better clarity.

### 2. Add ORDERNO field to sys_table (2026-01-12)
```bash
mysql -u root -p skyserver < sqls/migration_add_orderno_to_sys_table.sql
```

⚠️ **Required** if you have an existing database created before 2026-01-12.

This migration adds the `ORDERNO` field to the `sys_table` table to support menu sorting functionality.

## Archived Files

The `archived/` directory contains the original separate SQL files:
- `create_skyserver.sql` - Table definitions only
- `init_database.sql` - Old initialization script

These files are kept for reference but are no longer needed for setup.
