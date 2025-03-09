#!/bin/bash
# start-services.sh - Script to start the Medusa services

# Check if MongoDB is already running
echo "Checking if MongoDB is already running..."
if nc -z ${MONGODB_HOST:-localhost} ${MONGODB_PORT:-27017} 2>/dev/null; then
    echo "MongoDB is already running. Starting services without MongoDB..."
    
    # Update connection strings in environment files
    MONGODB_HOST=${MONGODB_HOST:-localhost}
    MONGODB_PORT=${MONGODB_PORT:-27017}
    
    # Start just the service containers (no MongoDB)
    docker-compose up -d
else
    echo "MongoDB is not running. Starting all services including MongoDB..."
    docker-compose --profile database up -d
fi

echo "Medusa services started successfully!"