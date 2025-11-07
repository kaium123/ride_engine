# GORM ORM Integration

The project now uses **GORM** (Go Object-Relational Mapping) instead of raw SQL queries for PostgreSQL operations.

## What Changed

### Before (Raw SQL)
```go
query := `SELECT id, phone, email FROM users WHERE id = $1`
err := db.QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Phone, &user.Email)
```

### After (GORM)
```go
var model UserModel
result := r.db.WithContext(ctx).Where("id = ?", id).First(&model)
```

## Benefits of GORM

✅ **Type Safety**: Compile-time checks for database operations
✅ **Auto Migrations**: Automatically creates and updates database schema
✅ **Associations**: Easy handling of relationships (foreign keys, joins)
✅ **Hooks**: Before/After create, update, delete callbacks
✅ **Query Builder**: Chainable methods for complex queries
✅ **Connection Pooling**: Built-in connection management
✅ **Transaction Support**: Easy transaction handling
✅ **Preloading**: Eager loading of associations

## Project Structure

```
internal/ride_engine/repository/postgres/
├── models.go              # GORM models (UserModel, DriverModel, RideModel)
├── user_postgres.go       # User repository with GORM
├── driver_postgres.go     # Driver repository with GORM
└── ride_postgres.go       # Ride repository with GORM
```

## GORM Models

### UserModel
```go
type UserModel struct {
    ID        string    `gorm:"type:varchar(255);primaryKey"`
    Phone     string    `gorm:"type:varchar(20);uniqueIndex;not null"`
    Email     string    `gorm:"type:varchar(255)"`
    Type      string    `gorm:"type:varchar(20);not null"`
    CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}
```

### DriverModel
```go
type DriverModel struct {
    ID            string     `gorm:"type:varchar(255);primaryKey"`
    IsOnline      bool       `gorm:"not null;default:false"`
    CurrentLat    *float64   `gorm:"type:double precision"`
    CurrentLng    *float64   `gorm:"type:double precision"`
    LastUpdatedAt *time.Time `gorm:"type:timestamp"`
    OTP           string     `gorm:"type:varchar(10)"`
    OTPExpiry     *time.Time `gorm:"type:timestamp"`
    User          UserModel  `gorm:"foreignKey:ID;references:ID;constraint:OnDelete:CASCADE"`
}
```

### RideModel
```go
type RideModel struct {
    ID          string     `gorm:"type:varchar(255);primaryKey"`
    RiderID     string     `gorm:"type:varchar(255);not null;index"`
    DriverID    *string    `gorm:"type:varchar(255);index"`
    PickupLat   float64    `gorm:"type:double precision;not null"`
    PickupLng   float64    `gorm:"type:double precision;not null"`
    DropoffLat  float64    `gorm:"type:double precision;not null"`
    DropoffLng  float64    `gorm:"type:double precision;not null"`
    Status      string     `gorm:"type:varchar(20);not null;index"`
    RequestedAt time.Time  `gorm:"not null;default:CURRENT_TIMESTAMP;index"`
    // ...timestamps
}
```

## Auto Migration

GORM automatically creates and updates the database schema on startup:

```go
// In main.go
if err := postgres.AutoMigrate(postgresDB.DB); err != nil {
    log.Fatalf("Failed to run migrations: %v", err)
}
```

This replaces the manual SQL scripts in `scripts/init-postgres.sql`.

## Common GORM Operations

### Create
```go
result := r.db.WithContext(ctx).Create(model)
```

### Find by ID
```go
result := r.db.WithContext(ctx).Where("id = ?", id).First(&model)
```

### Update
```go
result := r.db.WithContext(ctx).Model(&UserModel{}).
    Where("id = ?", user.ID).
    Updates(map[string]interface{}{
        "phone": model.Phone,
        "email": model.Email,
    })
```

### Delete
```go
result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&UserModel{})
```

### Find with conditions
```go
result := r.db.WithContext(ctx).
    Where("status IN ?", []string{"requested", "accepted"}).
    Order("requested_at DESC").
    Find(&models)
```

### Joins with Preload
```go
result := r.db.WithContext(ctx).
    Joins("JOIN users ON users.id = drivers.id").
    Where("users.phone = ?", phone).
    Preload("User").
    First(&model)
```

### Transactions
```go
return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(userModel).Error; err != nil {
        return err
    }
    if err := tx.Create(driverModel).Error; err != nil {
        return err
    }
    return nil
})
```

## Error Handling

### Record Not Found
```go
if errors.Is(result.Error, gorm.ErrRecordNotFound) {
    return nil, ErrUserNotFound
}
```

### Duplicate Key
```go
if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
    return ErrUserAlreadyExists
}
```

### Check Rows Affected
```go
if result.RowsAffected == 0 {
    return ErrUserNotFound
}
```

## Migration from Raw SQL

The migration maintains the same database schema and table names, ensuring backward compatibility. The SQL init scripts in `scripts/` can now be removed since GORM handles schema management.

## Configuration

GORM configuration in `pkg/database/postgres.go`:

```go
gormConfig := &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info), // SQL logging
    NowFunc: func() time.Time {
        return time.Now().UTC()
    },
    PrepareStmt: true, // Prepared statements for performance
}
```

## Performance Considerations

- **Prepared Statements**: Enabled by default for better performance
- **Connection Pooling**: Configured with `SetMaxOpenConns`, `SetMaxIdleConns`
- **Query Optimization**: GORM generates optimized SQL queries
- **Batch Operations**: Use `CreateInBatches` for bulk inserts
- **Select Specific Fields**: Use `Select()` to avoid loading unnecessary columns

## Testing

GORM makes testing easier:

```go
// Use SQLite for in-memory testing
db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

// Run migrations
db.AutoMigrate(&UserModel{}, &DriverModel{}, &RideModel{})

// Test repositories
repo := NewUserPostgresRepository(&database.PostgresDB{DB: db})
```

## Resources

- [GORM Documentation](https://gorm.io/docs/)
- [GORM PostgreSQL Driver](https://gorm.io/docs/connecting_to_the_database.html#PostgreSQL)
- [GORM Associations](https://gorm.io/docs/associations.html)
- [GORM Hooks](https://gorm.io/docs/hooks.html)
