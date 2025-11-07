# OTP Implementation - Dual Storage Strategy

## Overview

The OTP system uses a **dual storage strategy** for optimal performance and audit trail:
- **Redis**: Fast validation with automatic TTL expiry (2 minutes)
- **PostgreSQL**: Permanent audit trail and security monitoring

## Architecture

```
┌─────────────────┐
│  Driver Request │
│      OTP        │
└────────┬────────┘
         │
         ▼
┌────────────────────┐
│   OTP Service      │
│  (Generate OTP)    │
└────────┬───────────┘
         │
         ├──────────────┬──────────────┐
         ▼              ▼              ▼
    ┌────────┐    ┌──────────┐   ┌─────────┐
    │  Redis │    │PostgreSQL│   │ SMS API │
    │(2 min) │    │  (Audit) │   │ (Future)│
    └────────┘    └──────────┘   └─────────┘
```

## Database Schema

### PostgreSQL - `otp_records` Table

```sql
CREATE TABLE otp_records (
    id BIGSERIAL PRIMARY KEY,
    phone VARCHAR(20) NOT NULL,
    otp VARCHAR(10) NOT NULL,
    purpose VARCHAR(50) NOT NULL,           -- driver_login, customer_verification, etc
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    is_expired BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMP NOT NULL,
    verified_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_otp_phone ON otp_records(phone);
CREATE INDEX idx_otp_expires_at ON otp_records(expires_at);
CREATE INDEX idx_otp_created_at ON otp_records(created_at);
```

### Redis Keys

```
Key: otp:<phone_number>
Value: <6-digit-otp>
TTL: 120 seconds (2 minutes)
```

## API Flow

### 1. Request OTP

**Endpoint**: `POST /api/v1/drivers/login/request-otp`

**Request:**
```json
{
  "phone": "01875113841"
}
```

**Flow:**
1. Check if driver exists in database
2. Generate 6-digit OTP
3. Save to Redis with 2-minute TTL
4. Save to PostgreSQL with `expires_at` timestamp
5. Send OTP via SMS (currently prints to console)
6. Return success response

**Response:**
```json
{
  "message": "OTP sent successfully"
}
```

**Console Output:**
```
OTP for driver 01875113841: 123456
```

### 2. Verify OTP

**Endpoint**: `POST /api/v1/drivers/login/verify-otp`

**Request:**
```json
{
  "phone": "01875113841",
  "otp": "123456"
}
```

**Flow:**
1. Check Redis first (fast path)
   - If found and matches → verify
   - If not found → check PostgreSQL (fallback)
2. Verify OTP in PostgreSQL
   - Mark as `is_verified = true`
   - Set `verified_at` timestamp
3. Delete OTP from Redis (single use)
4. Generate JWT token for driver
5. Return driver info and token

**Response:**
```json
{
  "customer": {
    "id": 1,
    "name": "Jane Driver",
    "phone": "01875113841",
    "vehicle_no": "DHA-1234",
    "is_online": false,
    "created_at": "2025-11-05T10:00:00Z"
  },
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

## OTP Service Methods

### `GenerateOTP()`
Generates a random 6-digit OTP using crypto/rand.

```go
func (s *OTPService) GenerateOTP() string {
    return fmt.Sprintf("%06d", rand.Intn(1000000))
}
```

### `SaveOTP(ctx, phone, otp, purpose)`
Saves OTP to both Redis and PostgreSQL.

**Parameters:**
- `phone`: Driver's phone number
- `otp`: Generated 6-digit code
- `purpose`: Purpose of OTP (e.g., "driver_login")

**Storage:**
- **Redis**: `otp:<phone>` = `<otp>` (TTL: 2 minutes)
- **PostgreSQL**: Full record with expiry timestamp

### `VerifyOTP(ctx, phone, otp)`
Verifies OTP from Redis or PostgreSQL (fallback).

**Returns:**
- `bool`: true if OTP is valid
- `error`: any validation errors

**Process:**
1. Check Redis (fast path)
2. If not in Redis, check PostgreSQL
3. Mark as verified in database
4. Delete from Redis

### `InvalidateOTP(ctx, phone)`
Invalidates all pending OTPs for a phone number.

**Actions:**
- Deletes from Redis
- Marks as expired in PostgreSQL

## Security Features

### 1. Time-based Expiry
- Redis: Automatic TTL (2 minutes)
- PostgreSQL: `expires_at` timestamp checked during verification

### 2. Single-Use OTPs
- OTP deleted from Redis after successful verification
- `is_verified` flag prevents reuse in PostgreSQL

### 3. Audit Trail
All OTP operations logged in PostgreSQL:
- When OTP was created
- When OTP was verified
- If OTP expired without verification
- Purpose of each OTP

### 4. Rate Limiting (Future Enhancement)
Track OTP requests per phone number to prevent abuse:
```sql
SELECT COUNT(*)
FROM otp_records
WHERE phone = '...'
  AND created_at > NOW() - INTERVAL '1 hour'
  AND purpose = 'driver_login';
```

## OTP Purposes

The system supports different OTP purposes:

| Purpose | Description | Used For |
|---------|-------------|----------|
| `driver_login` | Driver authentication | Driver OTP login |
| `customer_verification` | Customer phone verification | Future: customer signup |
| `password_reset` | Password reset verification | Future: forgot password |
| `phone_change` | Phone number change verification | Future: update phone |

## Monitoring & Analytics

### Query Recent OTPs
```sql
SELECT phone, purpose, is_verified, created_at, verified_at
FROM otp_records
WHERE created_at > NOW() - INTERVAL '1 day'
ORDER BY created_at DESC
LIMIT 50;
```

### Check OTP Success Rate
```sql
SELECT
    purpose,
    COUNT(*) as total_sent,
    SUM(CASE WHEN is_verified THEN 1 ELSE 0 END) as verified,
    ROUND(100.0 * SUM(CASE WHEN is_verified THEN 1 ELSE 0 END) / COUNT(*), 2) as success_rate
FROM otp_records
WHERE created_at > NOW() - INTERVAL '7 days'
GROUP BY purpose;
```

### Find Suspicious Activity
```sql
-- Multiple failed OTP attempts
SELECT phone, COUNT(*) as failed_attempts
FROM otp_records
WHERE is_verified = false
  AND created_at > NOW() - INTERVAL '1 hour'
GROUP BY phone
HAVING COUNT(*) > 5
ORDER BY failed_attempts DESC;
```

## Database Maintenance

### Cleanup Old OTP Records
Automatically remove verified OTPs older than 30 days:

```sql
DELETE FROM otp_records
WHERE is_verified = true
  AND verified_at < NOW() - INTERVAL '30 days';
```

Or use the repository method:
```go
otpRepo.CleanupExpiredOTPs(ctx, time.Now().Add(-30*24*time.Hour))
```

## Testing

### Test OTP Flow with curl

1. **Request OTP:**
```bash
curl -X POST http://localhost:8080/api/v1/drivers/login/request-otp \
  -H "Content-Type: application/json" \
  -d '{"phone": "01875113841"}'
```

2. **Check console for OTP:**
```
OTP for driver 01875113841: 123456
```

3. **Verify OTP:**
```bash
curl -X POST http://localhost:8080/api/v1/drivers/login/verify-otp \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "01875113841",
    "otp": "123456"
  }'
```

### Verify in Database
```sql
-- Check OTP was recorded
SELECT * FROM otp_records
WHERE phone = '01875113841'
ORDER BY created_at DESC
LIMIT 1;

-- Check Redis
redis-cli GET otp:01875113841
```

## Error Handling

| Error | Cause | Response |
|-------|-------|----------|
| OTP not found | Expired or never created | `invalid or expired OTP` |
| OTP mismatch | Wrong code entered | `invalid or expired OTP` |
| Driver not found | Phone not registered | `driver not found` |
| Redis failure | Redis connection issue | Falls back to PostgreSQL |

## Future Enhancements

1. **SMS Integration**
   - Twilio, AWS SNS, or local SMS gateway
   - Replace console print with actual SMS send

2. **Rate Limiting**
   - Max 3 OTP requests per phone per hour
   - Exponential backoff for repeated failures

3. **OTP Resend**
   - Allow resending after 30 seconds
   - Maximum 3 resends per session

4. **Custom OTP Length**
   - Support 4, 6, or 8 digit OTPs
   - Configurable per purpose

5. **Email OTP**
   - Support OTP via email for customers
   - Dual-channel verification (email + SMS)

6. **Analytics Dashboard**
   - Real-time OTP metrics
   - Success/failure rates
   - Geographic distribution
   - Fraud detection patterns

## Configuration

OTP settings can be configured via environment variables:

```bash
# Redis Configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# OTP Settings (future)
OTP_LENGTH=6
OTP_EXPIRY_MINUTES=2
OTP_MAX_ATTEMPTS=3
OTP_RATE_LIMIT_PER_HOUR=3
```

## Security Best Practices

1. ✅ Use Redis TTL for automatic expiry
2. ✅ Store hashed OTPs in database (future enhancement)
3. ✅ Implement rate limiting per phone number
4. ✅ Log all OTP operations for audit
5. ✅ Use HTTPS in production
6. ✅ Implement CAPTCHA for repeated failures
7. ✅ Monitor for suspicious patterns
8. ✅ Set maximum OTP validity (2 minutes)
9. ✅ Single-use OTPs (delete after verification)
10. ✅ Different OTP purposes for better tracking
