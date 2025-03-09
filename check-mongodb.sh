#!/bin/bash
# check-mongodb.sh - Script to check if MongoDB is already running

# Default MongoDB port
PORT=${MONGODB_PORT:-27017}
# Default MongoDB host
HOST=${MONGODB_HOST:-localhost}

# Check if MongoDB is already running on the specified port
if nc -z $HOST $PORT 2>/dev/null; then
    echo "MongoDB is already running on $HOST:$PORT"
    echo "MONGODB_RUNNING=true" > .mongodb-status
else
    echo "MongoDB is not running on $HOST:$PORT"
    echo "MONGODB_RUNNING=false" > .mongodb-status
fi