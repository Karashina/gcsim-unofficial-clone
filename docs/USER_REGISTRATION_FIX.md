# User Registration System - Email Field Removal

**Last Updated**: 2026-01-04  
**Status**: Production Ready  
**Breaking Change**: Yes (email field removed)

## Overview

This document describes the removal of the email field from the user registration system to resolve UNIQUE constraint issues that prevented multiple users from registering.

## Problem Statement

### Symptoms
- User registration failed with HTTP 500 error after the first user
- Server logs showed: `UNIQUE constraint failed: users.email`
- Only one user could register successfully

### Root Cause
The `users` table had a UNIQUE constraint on the `email` column. Since the registration form did not require email input, all users had NULL or empty email values, causing SQLite's UNIQUE constraint to fail on the second registration attempt.

**Important Note**: Unlike PostgreSQL, SQLite's UNIQUE constraint does NOT allow multiple NULL values.

## Solution

Complete removal of the email field from the entire codebase, including:
- Database schema
- User model structures
- API request/response objects
- Validation logic
- Repository methods

## Technical Changes

### 1. Database Schema

#### Before
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    email TEXT,  -- ❌ REMOVED
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    status TEXT NOT NULL DEFAULT 'pending',
    ...
)
CREATE UNIQUE INDEX idx_users_email ON users(email);  -- ❌ REMOVED
```

#### After
```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    status TEXT NOT NULL DEFAULT 'pending',
    created_at DATETIME,
    updated_at DATETIME,
    approved_at DATETIME,
    approved_by INTEGER
)
CREATE UNIQUE INDEX idx_users_username ON users(username);  -- ✅ Kept
```

### 2. Go Structures

#### User Model (`internal/db/user.go`)
```go
// Before
type User struct {
    ID           uint
    Username     string
    Email        *string  // ❌ REMOVED
    PasswordHash string
    ...
}

// After
type User struct {
    ID           uint
    Username     string
    PasswordHash string
    Role         UserRole
    Status       UserStatus
    CreatedAt    time.Time
    UpdatedAt    time.Time
    ApprovedAt   *time.Time
    ApprovedBy   *uint
}
```

#### Registration Request (`internal/auth/handlers.go`)
```go
// Before
type RegisterRequest struct {
    Username string `json:"username"`
    Email    string `json:"email"`  // ❌ REMOVED
    Password string `json:"password"`
}

// After
type RegisterRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}
```

### 3. Removed Functions

- `FindByEmail(email string) (*User, error)` - No longer needed
- `validateEmail(email string) bool` - Email validation removed
- `ensureEmailIndexIsNotUnique()` - Index management removed

### 4. Updated Function Signatures

#### InitializeAdminUser
```go
// Before
func (d *Database) InitializeAdminUser(username, password, email string) error

// After
func (d *Database) InitializeAdminUser(username, password string) error
```

## Migration Strategy

### Automatic Migration
The `migrateUserTable()` function now automatically:
1. Drops the `email` column if it exists
2. Drops all email-related indexes
3. Creates the new schema without email field

```go
func (d *Database) migrateUserTable() error {
    // Drop email column if it exists
    d.db.Exec("ALTER TABLE users DROP COLUMN email")
    
    // Drop email-related indexes
    d.db.Exec("DROP INDEX IF EXISTS idx_users_email")
    d.db.Exec("DROP INDEX IF EXISTS idx_users_email_unique")
    
    return d.db.AutoMigrate(&User{})
}
```

### Manual Migration (if needed)
If automatic migration fails:

```bash
# Backup existing database
sudo cp /var/www/html/data/gcsim.db /var/www/html/data/gcsim.db.backup

# Stop service
sudo systemctl stop gcsim-webui

# Remove database (loses all data!)
sudo rm /var/www/html/data/gcsim.db

# Start service (creates new database with correct schema)
sudo systemctl start gcsim-webui
```

## API Changes

### Registration Endpoint

**Endpoint**: `POST /api/register`

#### Request Body (Before)
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "SecurePass123!"
}
```

#### Request Body (After)
```json
{
  "username": "testuser",
  "password": "SecurePass123!"
}
```

#### Validation Rules
- **Username**:
  - Required
  - 3-50 characters
  - Alphanumeric, underscore, and hyphen only
  - Must be unique
- **Password**:
  - Required
  - Minimum 8 characters
  - Must contain at least 2 of: uppercase, lowercase, digits, special characters

### Response
```json
{
  "success": true,
  "message": "ユーザー登録が完了しました。管理者の承認をお待ちください。"
}
```

## Verification

### Check Database Schema
```bash
sudo strings /var/www/html/data/gcsim.db | grep "CREATE TABLE.*users" -A10
```

Expected output (no email field):
```
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'user',
    status TEXT NOT NULL DEFAULT 'pending',
    ...
)
```

### Check for Email References
```bash
sudo strings /var/www/html/data/gcsim.db | grep -i "email"
```

Expected: No matches (exit code 1)

### Test User Registration
```bash
# Register first user
curl -X POST http://localhost:8382/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser1","password":"TestPass123!"}'

# Register second user (should succeed now)
curl -X POST http://localhost:8382/api/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser2","password":"TestPass123!"}'
```

Both registrations should return success messages.

## Impact Assessment

### ✅ Benefits
1. **Simplified registration flow** - No unnecessary fields
2. **Resolved UNIQUE constraint issue** - Multiple users can register
3. **Reduced code complexity** - ~100+ lines of code removed
4. **Better data model** - Only required fields present
5. **Improved maintainability** - Fewer fields to manage

### ⚠️ Trade-offs
1. **No email-based password reset** - Not implemented anyway
2. **No email notifications** - Not implemented anyway
3. **No email-based user lookup** - Username lookup is sufficient
4. **Breaking change** - Existing API contracts changed

### 🔍 No Impact
- Existing users unaffected (email data was not used)
- Login flow unchanged (uses username + password)
- Admin approval workflow unchanged
- JWT authentication unchanged

## Lessons Learned

### GORM Limitations
1. **AutoMigrate() doesn't remove fields** - Only adds new ones
2. **UNIQUE constraints persist** - Manual removal required
3. **Pointer types don't bypass constraints** - SQLite enforces UNIQUE on NULL
4. **Migration timing matters** - AutoMigrate() can override manual changes

### SQLite vs PostgreSQL
- **SQLite**: UNIQUE constraint rejects multiple NULLs
- **PostgreSQL**: UNIQUE constraint allows multiple NULLs
- Always test database-specific behavior

### Problem-Solving Approach
1. **Root cause over workarounds** - Removing unused features is better than complex fixes
2. **YAGNI principle** - Don't implement features "just in case"
3. **Verify deployments** - Always check binary timestamps and schema after deployment

## Future Considerations

### If Email is Needed Later

Should email functionality be required in the future:

1. **Make it optional** - Allow NULL values
2. **Don't use UNIQUE constraint** - Use application-level validation instead
3. **Store in separate table** - Keep core user table simple
4. **Use verification flags** - Track whether email is verified

Example future schema:
```sql
CREATE TABLE user_emails (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL,
    email TEXT NOT NULL,
    verified BOOLEAN DEFAULT FALSE,
    primary_email BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id)
)
CREATE INDEX idx_user_emails_user_id ON user_emails(user_id);
-- Note: No UNIQUE constraint on email!
```

### Recommended Next Steps

1. **Add integration tests** - Test user registration flow
2. **Monitor registration errors** - Set up alerts
3. **Document user management** - Admin workflows
4. **Review deploy scripts** - Ensure reliable binary updates

## References

- [AUTHENTICATION.md](AUTHENTICATION.md) - JWT authentication system
- [LOCAL_DEVELOPMENT.md](LOCAL_DEVELOPMENT.md) - Development setup
- [Work Log: 20260104-230000](../.github/work-logs/20260104-230000-remove-email-field-from-user-model.md) - Detailed implementation log

## Support

For questions or issues:
1. Check server logs: `sudo journalctl -u gcsim-webui -n 50`
2. Verify database schema: `sudo strings /var/www/html/data/gcsim.db | grep users`
3. Review this document and related files

---

**Document Version**: 1.0  
**Implementation Date**: 2026-01-04  
**Tested**: ✅ Production  
**Status**: Active
