#!/bin/bash

# Test database connections for Ride Engine
# This script verifies PostgreSQL, MongoDB, and Redis are accessible

echo "======================================"
echo "  Ride Engine - Database Connection Test"
echo "======================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test PostgreSQL
echo -n "Testing PostgreSQL (port 5436)... "
if command -v psql &> /dev/null; then
    if PGPASSWORD=secret psql -h localhost -p 5436 -U root -d ride_engine -c "SELECT 1" &> /dev/null; then
        echo -e "${GREEN}✓ Connected${NC}"

        # Count tables
        TABLE_COUNT=$(PGPASSWORD=secret psql -h localhost -p 5436 -U root -d ride_engine -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_type = 'BASE TABLE';" 2>/dev/null | tr -d ' ')
        echo "  - Tables found: $TABLE_COUNT"

        # Count customers
        CUSTOMER_COUNT=$(PGPASSWORD=secret psql -h localhost -p 5436 -U root -d ride_engine -t -c "SELECT COUNT(*) FROM customers;" 2>/dev/null | tr -d ' ')
        echo "  - Customers: $CUSTOMER_COUNT"

        # Count drivers
        DRIVER_COUNT=$(PGPASSWORD=secret psql -h localhost -p 5436 -U root -d ride_engine -t -c "SELECT COUNT(*) FROM drivers;" 2>/dev/null | tr -d ' ')
        echo "  - Drivers: $DRIVER_COUNT"

        # Count rides
        RIDE_COUNT=$(PGPASSWORD=secret psql -h localhost -p 5436 -U root -d ride_engine -t -c "SELECT COUNT(*) FROM rides;" 2>/dev/null | tr -d ' ')
        echo "  - Rides: $RIDE_COUNT"
    else
        echo -e "${RED}✗ Connection failed${NC}"
        echo "  - Ensure PostgreSQL is running on port 5436"
        echo "  - Try: docker-compose up -d postgres"
    fi
else
    echo -e "${YELLOW}⚠ psql not found${NC}"
    echo "  - Install PostgreSQL client or use Docker: docker exec -it ride_engine-postgres psql -U root -d ride_engine"
fi

echo ""

# Test MongoDB
echo -n "Testing MongoDB (port 27016)... "
if command -v mongosh &> /dev/null; then
    if mongosh "mongodb://root:secret@localhost:27016/ride_engine?authSource=admin" --quiet --eval "db.runCommand({ ping: 1 })" &> /dev/null; then
        echo -e "${GREEN}✓ Connected${NC}"

        # Count collections
        COLLECTION_COUNT=$(mongosh "mongodb://root:secret@localhost:27016/ride_engine?authSource=admin" --quiet --eval "db.getCollectionNames().length" 2>/dev/null)
        echo "  - Collections: $COLLECTION_COUNT"

        # Count driver locations
        LOCATION_COUNT=$(mongosh "mongodb://root:secret@localhost:27016/ride_engine?authSource=admin" --quiet --eval "db.driver_locations.countDocuments()" 2>/dev/null)
        echo "  - Driver locations: $LOCATION_COUNT"

        # Check indexes
        INDEX_COUNT=$(mongosh "mongodb://root:secret@localhost:27016/ride_engine?authSource=admin" --quiet --eval "db.driver_locations.getIndexes().length" 2>/dev/null)
        echo "  - Indexes on driver_locations: $INDEX_COUNT"
    else
        echo -e "${RED}✗ Connection failed${NC}"
        echo "  - Ensure MongoDB is running on port 27016"
        echo "  - Try: docker-compose up -d mongodb"
    fi
elif command -v mongo &> /dev/null; then
    # Fallback to legacy mongo shell
    if mongo "mongodb://root:secret@localhost:27016/ride_engine?authSource=admin" --quiet --eval "db.runCommand({ ping: 1 })" &> /dev/null; then
        echo -e "${GREEN}✓ Connected (using legacy mongo shell)${NC}"
    else
        echo -e "${RED}✗ Connection failed${NC}"
    fi
else
    echo -e "${YELLOW}⚠ mongosh/mongo not found${NC}"
    echo "  - Install MongoDB shell or use Docker: docker exec -it ride_engine-mongo mongosh"
fi

echo ""

# Test Redis
echo -n "Testing Redis (port 6379)... "
if command -v redis-cli &> /dev/null; then
    if redis-cli -h localhost -p 6379 PING &> /dev/null; then
        PONG=$(redis-cli -h localhost -p 6379 PING 2>/dev/null)
        echo -e "${GREEN}✓ Connected (${PONG})${NC}"

        # Count keys
        KEY_COUNT=$(redis-cli -h localhost -p 6379 DBSIZE 2>/dev/null | awk '{print $2}')
        echo "  - Keys in DB: $KEY_COUNT"

        # Memory usage
        MEMORY=$(redis-cli -h localhost -p 6379 INFO memory 2>/dev/null | grep "used_memory_human" | cut -d: -f2 | tr -d '\r')
        echo "  - Memory used: $MEMORY"
    else
        echo -e "${RED}✗ Connection failed${NC}"
        echo "  - Ensure Redis is running on port 6379"
        echo "  - Try: docker-compose up -d redis"
    fi
else
    echo -e "${YELLOW}⚠ redis-cli not found${NC}"
    echo "  - Install Redis client or use Docker: docker exec -it ride_engine-redis redis-cli"
fi

echo ""
echo "======================================"
echo "  Connection test completed"
echo "======================================"
echo ""

# Summary
echo "Next steps:"
echo "  1. If all connections passed, run: go run cmd/api/main.go"
echo "  2. API will be available at: http://localhost:8080"
echo "  3. Check endpoints: /health for health check"
echo ""
