#!/bin/bash

# Check if the application is running
curl -s http://localhost:8080/health > /dev/null
app_status=$?

# Check if Redis server is reachable
redis-cli PING > /dev/null
redis_status=$?

if [ $app_status -eq 0 ]; then
    echo "Application is running"
else
    echo "Application is not running"
fi

if [ $redis_status -eq 0 ]; then
    echo "Redis server is reachable"
else
    echo "Redis server is not reachable"
fi