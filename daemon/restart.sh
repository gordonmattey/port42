#!/bin/bash
# Proper daemon restart script with cleanup

echo "🔄 Restarting Port42 daemon..."

# Step 1: Kill any hanging port42 CLI processes (they might be holding connections)
echo "  Cleaning up CLI processes..."
pkill -f "port42 possess" 2>/dev/null
pkill -f "port42 shell" 2>/dev/null
pkill -f "port42 context --watch" 2>/dev/null

# Step 2: Kill the daemon gracefully
echo "  Stopping daemon..."
pkill -TERM port42-daemon 2>/dev/null || pkill -TERM port42d 2>/dev/null
sleep 1

# Step 3: Force kill if still running
if pgrep -f port42-daemon > /dev/null || pgrep -f port42d > /dev/null; then
    echo "  Force stopping daemon..."
    pkill -KILL port42-daemon 2>/dev/null
    pkill -KILL port42d 2>/dev/null
    sleep 1
fi

# Step 4: Build
echo "  Building daemon..."
cd /Users/gordon/Dropbox/Work/Hacking/workspace/port42
./build.sh

# Step 5: Install (copy to bin)
echo "  Installing daemon..."
cp bin/port42d /Users/gordon/.port42/bin/ 2>/dev/null || true
cp bin/port42 /Users/gordon/.port42/bin/ 2>/dev/null || true

# Step 6: Start fresh
echo "  Starting daemon..."
/Users/gordon/.port42/bin/port42d &

echo "✅ Daemon restarted successfully"