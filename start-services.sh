#!/bin/bash
# start-services.sh - Script to start the Medusa services

# Check if MongoDB is already running
echo "Checking if MongoDB is already running..."
./check-mongodb.sh

# Source the MongoDB status
source .mongodb-status

# Determine which profile to use
if [ "$MONGODB_RUNNING" = "true" ]; then
    echo "Using existing MongoDB instance..."
    
    # Update the environment files to use the existing MongoDB
    MONGODB_HOST=${MONGODB_HOST:-localhost}
    MONGODB_PORT=${MONGODB_PORT:-27017}
    
    # Update the Auth Service .env file
    if [ -f ./auth-service/.env ]; then
        sed -i "s|MONGO_URI=.*|MONGO_URI=mongodb://$MONGODB_HOST:$MONGODB_PORT|g" ./auth-service/.env
    fi
    
    # Update the Core Service .env file
    if [ -f ./core-service/.env ]; then
        sed -i "s|MONGO_URI=.*|MONGO_URI=mongodb://$MONGODB_HOST:$MONGODB_PORT|g" ./core-service/.env
    fi
    
    # Start just the service containers (no MongoDB)
    docker-compose up -d auth-service core-service
else
    echo "Starting all services including MongoDB..."
    docker-compose up -d
fi

echo "Medusa services started successfully!"