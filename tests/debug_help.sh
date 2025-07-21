#!/bin/bash
# Debug which help is being shown

echo "=== Debugging Port 42 Help ==="
echo

# Check binary info
echo "1. Binary locations and checksums:"
for binary in /usr/local/bin/port42 ./bin/port42 ./cli/target/release/port42; do
    if [ -f "$binary" ]; then
        echo -n "$binary: "
        shasum -a 256 "$binary" | cut -d' ' -f1
    fi
done

echo
echo "2. Checking for 'memory list' string in binaries:"
for binary in /usr/local/bin/port42 ./bin/port42 ./cli/target/release/port42; do
    if [ -f "$binary" ]; then
        echo -n "$binary: "
        if strings "$binary" | grep -q "memory list"; then
            echo "FOUND"
        else
            echo "NOT FOUND"
        fi
    fi
done

echo
echo "3. Shell help test (type 'help' then 'exit'):"
echo "================================================"