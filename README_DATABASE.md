# Database Setup Guide

This guide explains how to set up and use PostgreSQL and MongoDB for the Ride Engine project.

## Architecture

The Ride Engine uses a **polyglot persistence** approach:

- **PostgreSQL**: Stores structured data (users, drivers, rides)
- **MongoDB**: Stores location tracking data with geospatial queries

## Quick Start

### 1. Start Databases with Docker

```bash
# Start all services
docker-compose up -d

# Or use the Makefile
make docker-up
```

This will start:
- PostgreSQL on port `5432`
- MongoDB on port `27017`
- pgAdmin (optional) on port `5050`
- Mongo Express (optional) on port `8081`

### 2. Verify Databases are Running

```bash
# Check container status
docker-compose ps

# View logs
docker-compose logs -f postgres
docker-compose logs -f mongodb
```

### 3. Run the Application

```bash
# Run directly
go run cmd/api/main.go

# Or build and run
make build
./bin/ride_engine
```

## Database Credentials

### PostgreSQL
- **Host**: localhost
- **Port**: 5432
- **Database**: ride_engine_db
- **User**: ride_engine
- **Password**: ride_engine_password

### MongoDB
- **URI**: mongodb://ride_engine:ride_engine_password@localhost:27017
- **Database**: ride_engine_locations

## PostgreSQL Schema

### Tables

#### `users`
Stores all users (both riders and drivers)
```sql
- id (VARCHAR, PRIMARY KEY)
- phone (VARCHAR, UNIQUE)
- email (VARCHAR)
- type (VARCHAR: 'rider' or 'driver')
- created_at (TIMESTAMP)
```

#### `drivers`
Stores driver-specific information
```sql
- id (VARCHAR, FOREIGN KEY -> users.id)
- is_online (BOOLEAN)
- current_lat (DOUBLE PRECISION)
- current_lng (DOUBLE PRECISION)
- last_updated_at (TIMESTAMP)
- otp (VARCHAR)
- otp_expiry (TIMESTAMP)
```

#### `rides`
Stores ride requests and tracking
```sql
- id (VARCHAR, PRIMARY KEY)
- rider_id (VARCHAR, FOREIGN KEY -> users.id)
- driver_id (VARCHAR, FOREIGN KEY -> users.id)
- pickup_lat, pickup_lng (DOUBLE PRECISION)
- dropoff_lat, dropoff_lng (DOUBLE PRECISION)
- status (VARCHAR)
- requested_at, accepted_at, started_at, completed_at, cancelled_at (TIMESTAMP)
```

## MongoDB Collections

### `driver_locations`
Stores historical location data for drivers
```javascript
{
  driver_id: String,
  location: {
    type: "Point",
    coordinates: [longitude, latitude]  // GeoJSON format
  },
  timestamp: Date,
  is_online: Boolean,
  metadata: Object (optional)
}
```

**Indexes**:
- 2dsphere index on `location` for geospatial queries
- Compound index on `driver_id` and `timestamp`
- TTL index on `timestamp` (auto-deletes after 30 days)

### `ride_locations`
Stores location tracking during rides
```javascript
{
  ride_id: String,
  driver_id: String,
  location: {
    type: "Point",
    coordinates: [longitude, latitude]
  },
  timestamp: Date,
  status: String
}
```

**Indexes**:
- 2dsphere index on `location`
- Compound indexes on `ride_id` and `driver_id`

## Management Tools

### pgAdmin (PostgreSQL GUI)
- **URL**: http://localhost:5050
- **Email**: admin@rideengine.com
- **Password**: admin

To connect to PostgreSQL from pgAdmin:
1. Add new server
2. Host: postgres (or localhost if connecting from host machine)
3. Port: 5432
4. Database: ride_engine_db
5. Username: ride_engine
6. Password: ride_engine_password

### Mongo Express (MongoDB GUI)
- **URL**: http://localhost:8081
- **Username**: admin
- **Password**: admin

## Common Operations

### Reset Databases
```bash
# Stop containers and remove volumes
docker-compose down -v

# Start fresh
docker-compose up -d

# Or use Makefile
make db-reset
```

### Access Database Shells

#### PostgreSQL
```bash
docker exec -it ride_engine_postgres psql -U ride_engine -d ride_engine_db
```

Common psql commands:
```sql
\dt                          -- List tables
\d users                     -- Describe users table
SELECT * FROM users LIMIT 5; -- Query data
```

#### MongoDB
```bash
docker exec -it ride_engine_mongodb mongosh -u ride_engine -p ride_engine_password
```

Common mongosh commands:
```javascript
use ride_engine_locations
show collections
db.driver_locations.find().limit(5)
db.driver_locations.countDocuments()
```

### Backup and Restore

#### PostgreSQL Backup
```bash
docker exec -t ride_engine_postgres pg_dump -U ride_engine ride_engine_db > backup.sql
```

#### PostgreSQL Restore
```bash
docker exec -i ride_engine_postgres psql -U ride_engine -d ride_engine_db < backup.sql
```

#### MongoDB Backup
```bash
docker exec ride_engine_mongodb mongodump --username ride_engine --password ride_engine_password --authenticationDatabase admin --db ride_engine_locations --out /tmp/backup
docker cp ride_engine_mongodb:/tmp/backup ./mongo_backup
```

#### MongoDB Restore
```bash
docker cp ./mongo_backup ride_engine_mongodb:/tmp/backup
docker exec ride_engine_mongodb mongorestore --username ride_engine --password ride_engine_password --authenticationDatabase admin /tmp/backup
```

## Geospatial Queries

### Find Nearby Drivers
```javascript
// In MongoDB
db.driver_locations.find({
  location: {
    $near: {
      $geometry: {
        type: "Point",
        coordinates: [90.4125, 23.8103]  // [lng, lat] - Dhaka
      },
      $maxDistance: 5000  // 5km in meters
    }
  },
  is_online: true
})
```

## Environment Variables

Create a `.env` file based on `.env.example`:

```bash
cp .env.example .env
```

Edit the `.env` file if you need to change default configurations.

## Production Considerations

For production deployment:

1. **Use strong passwords** - Change default credentials
2. **Enable SSL/TLS** - Set `POSTGRES_SSLMODE=require`
3. **Set up replication** - For high availability
4. **Configure backups** - Regular automated backups
5. **Monitor performance** - Use monitoring tools
6. **Resource limits** - Configure appropriate memory/CPU limits
7. **Network security** - Use private networks, not expose ports publicly
8. **Data retention** - Adjust MongoDB TTL index as needed

## Troubleshooting

### PostgreSQL Connection Issues
```bash
# Check if PostgreSQL is running
docker-compose ps postgres

# View PostgreSQL logs
docker-compose logs postgres

# Test connection
docker exec -it ride_engine_postgres pg_isready -U ride_engine
```

### MongoDB Connection Issues
```bash
# Check if MongoDB is running
docker-compose ps mongodb

# View MongoDB logs
docker-compose logs mongodb

# Test connection
docker exec -it ride_engine_mongodb mongosh --eval "db.adminCommand('ping')"
```

### Port Already in Use
If ports 5432 or 27017 are already in use, edit `docker-compose.yml` to use different ports:

```yaml
ports:
  - "5433:5432"  # Use 5433 on host instead of 5432
```

Then update your `.env` file accordingly.

## Additional Resources

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [MongoDB Documentation](https://docs.mongodb.com/)
- [PostGIS (Geospatial for PostgreSQL)](https://postgis.net/)
- [MongoDB Geospatial Queries](https://docs.mongodb.com/manual/geospatial-queries/)
