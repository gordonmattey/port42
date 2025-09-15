#!/bin/bash
# Test watch mode display
(
    sleep 1
    echo "Starting watch mode test..."
    sleep 2
    # Send Ctrl+C after 3 seconds
    kill -INT $$
) &

port42 context --watch
