# Scripts Directory

This directory is for database maintenance and migration scripts.

## Directory Structure

```
scripts/
├── README.md          # This file
└── archived/          # Archived one-time migration scripts
```

## Archived Scripts

The `archived/` directory contains historical one-time scripts that were used during development:

### Permission Setup Scripts
- `setup_permissions.go` - Initial permission system setup
- `setup_perms_final.go` - Final permission configuration
- `setup_perms_simple.go` - Simplified permission setup
- `fix_permissions.go` - Permission bug fixes

### Hook Management Scripts
- `add_sys_table_hook.go` - Add sys_table create hook
- `check_hooks.go` - Verify hook configuration
- `fix_hook_name.go` - Fix hook naming issues
- `remove_duplicate_hook.go` - Clean up duplicate hooks
- `verify_hook_success.go` - Verify hook execution

### Column Migration Scripts
- `add_orderno_column.go` - Add ORDERNO column
- `add_orderno_metadata.go` - Add ORDERNO to metadata
- `add_orderno_to_all.go` - Batch add ORDERNO to all tables

## Notes

- All scripts in `archived/` have been executed and are kept for reference only
- Do not run archived scripts on production databases
- New migration scripts should be created in this directory when needed
- Consider using proper migration tools for production environments
