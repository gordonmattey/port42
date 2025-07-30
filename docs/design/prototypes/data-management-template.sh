#!/bin/bash
# Port 42 Data Management Command Template
# This template shows how to create CRUD commands for structured data

# Example: content-calendar command
# Manages blog posts, social media content, and publication schedules

set -e

COMMAND_NAME="content-calendar"
DATA_FILE="$HOME/.port42/data/${COMMAND_NAME}.json"
DATA_DIR="$(dirname "$DATA_FILE")"

# Ensure data directory exists
mkdir -p "$DATA_DIR"

# Initialize data file if it doesn't exist
if [ ! -f "$DATA_FILE" ]; then
    echo '{"entries": []}' > "$DATA_FILE"
fi

# Color codes for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Helper functions
print_usage() {
    cat << EOF
Usage: $COMMAND_NAME <command> [options]

Commands:
  create    Add a new content entry
  list      List all content entries
  show      Show details of a specific entry
  update    Update an existing entry
  delete    Delete an entry
  schedule  View content schedule
  stats     Show content statistics

Options:
  -t, --type <type>      Content type (blog, social, video)
  -s, --status <status>  Status (draft, scheduled, published)
  -d, --date <date>      Publication date (YYYY-MM-DD)
  -i, --id <id>          Entry ID
  --json                 Output in JSON format

Examples:
  $COMMAND_NAME create -t blog "How to Use Port 42"
  $COMMAND_NAME list -s draft
  $COMMAND_NAME update -i 123 -s published
  $COMMAND_NAME schedule --date 2024-01-15

EOF
}

# Generate unique ID
generate_id() {
    echo "$(date +%s)$(shuf -i 100-999 -n 1)"
}

# Create new entry
cmd_create() {
    local title=""
    local type="blog"
    local status="draft"
    local date=$(date +%Y-%m-%d)
    local tags=""
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--type) type="$2"; shift 2 ;;
            -s|--status) status="$2"; shift 2 ;;
            -d|--date) date="$2"; shift 2 ;;
            --tags) tags="$2"; shift 2 ;;
            *) title="$1"; shift ;;
        esac
    done
    
    if [ -z "$title" ]; then
        echo -e "${RED}Error: Title is required${NC}"
        echo "Usage: $COMMAND_NAME create [options] <title>"
        exit 1
    fi
    
    # Create entry
    local id=$(generate_id)
    local entry=$(jq -n \
        --arg id "$id" \
        --arg title "$title" \
        --arg type "$type" \
        --arg status "$status" \
        --arg date "$date" \
        --arg created "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
        --arg tags "$tags" \
        '{
            id: $id,
            title: $title,
            type: $type,
            status: $status,
            date: $date,
            created: $created,
            updated: $created,
            tags: ($tags | split(",") | map(ltrimstr(" ") | rtrimstr(" "))),
            metadata: {}
        }')
    
    # Add to data file
    jq ".entries += [$entry]" "$DATA_FILE" > "$DATA_FILE.tmp" && mv "$DATA_FILE.tmp" "$DATA_FILE"
    
    echo -e "${GREEN}✓ Created content entry:${NC}"
    echo -e "  ID: ${BLUE}$id${NC}"
    echo -e "  Title: $title"
    echo -e "  Type: $type"
    echo -e "  Status: $status"
    echo -e "  Date: $date"
}

# List entries
cmd_list() {
    local filter_type=""
    local filter_status=""
    local json_output=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            -t|--type) filter_type="$2"; shift 2 ;;
            -s|--status) filter_status="$2"; shift 2 ;;
            --json) json_output=true; shift ;;
            *) shift ;;
        esac
    done
    
    # Build jq filter
    local jq_filter=".entries"
    if [ -n "$filter_type" ]; then
        jq_filter="$jq_filter | map(select(.type == \"$filter_type\"))"
    fi
    if [ -n "$filter_status" ]; then
        jq_filter="$jq_filter | map(select(.status == \"$filter_status\"))"
    fi
    
    if [ "$json_output" = true ]; then
        jq "$jq_filter" "$DATA_FILE"
    else
        # Pretty print
        echo -e "${BLUE}Content Calendar Entries:${NC}"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        
        jq -r "$jq_filter | .[] | 
            \"ID: \(.id) | \(.date) | \(.type) | \(.status)\n  ↳ \(.title)\"" "$DATA_FILE" | \
        while IFS= read -r line; do
            if [[ $line == ID:* ]]; then
                echo -e "${YELLOW}$line${NC}"
            else
                echo "$line"
            fi
        done
        
        local count=$(jq "$jq_filter | length" "$DATA_FILE")
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo -e "Total: ${GREEN}$count${NC} entries"
    fi
}

# Show entry details
cmd_show() {
    local id=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -i|--id) id="$2"; shift 2 ;;
            *) id="$1"; shift ;;
        esac
    done
    
    if [ -z "$id" ]; then
        echo -e "${RED}Error: Entry ID required${NC}"
        exit 1
    fi
    
    local entry=$(jq ".entries[] | select(.id == \"$id\")" "$DATA_FILE")
    if [ -z "$entry" ]; then
        echo -e "${RED}Error: Entry not found${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}Content Entry Details:${NC}"
    echo "$entry" | jq .
}

# Update entry
cmd_update() {
    local id=""
    local updates=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -i|--id) id="$2"; shift 2 ;;
            -t|--type) updates="$updates | .type = \"$2\""; shift 2 ;;
            -s|--status) updates="$updates | .status = \"$2\""; shift 2 ;;
            -d|--date) updates="$updates | .date = \"$2\""; shift 2 ;;
            --title) updates="$updates | .title = \"$2\""; shift 2 ;;
            *) shift ;;
        esac
    done
    
    if [ -z "$id" ]; then
        echo -e "${RED}Error: Entry ID required${NC}"
        exit 1
    fi
    
    # Update timestamp
    updates="$updates | .updated = \"$(date -u +%Y-%m-%dT%H:%M:%SZ)\""
    
    # Apply updates
    jq "(.entries[] | select(.id == \"$id\")) $updates" "$DATA_FILE" > "$DATA_FILE.tmp" && \
        mv "$DATA_FILE.tmp" "$DATA_FILE"
    
    echo -e "${GREEN}✓ Updated entry $id${NC}"
}

# Delete entry
cmd_delete() {
    local id=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -i|--id) id="$2"; shift 2 ;;
            *) id="$1"; shift ;;
        esac
    done
    
    if [ -z "$id" ]; then
        echo -e "${RED}Error: Entry ID required${NC}"
        exit 1
    fi
    
    # Confirm deletion
    echo -e "${YELLOW}Are you sure you want to delete entry $id? (y/N)${NC}"
    read -r confirm
    if [[ ! "$confirm" =~ ^[Yy]$ ]]; then
        echo "Deletion cancelled"
        exit 0
    fi
    
    jq ".entries = [.entries[] | select(.id != \"$id\")]" "$DATA_FILE" > "$DATA_FILE.tmp" && \
        mv "$DATA_FILE.tmp" "$DATA_FILE"
    
    echo -e "${GREEN}✓ Deleted entry $id${NC}"
}

# View schedule
cmd_schedule() {
    local date_filter=""
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            -d|--date) date_filter="$2"; shift 2 ;;
            *) shift ;;
        esac
    done
    
    echo -e "${BLUE}Content Schedule:${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    # Group by date
    if [ -n "$date_filter" ]; then
        jq -r ".entries[] | select(.date == \"$date_filter\") | 
            \"\(.date) | \(.type) | \(.status) | \(.title)\"" "$DATA_FILE"
    else
        jq -r '.entries | sort_by(.date) | group_by(.date) | 
            map("\(.[0].date):\n" + (map("  • [\(.type)] \(.title) (\(.status))") | join("\n"))) | 
            join("\n\n")' "$DATA_FILE"
    fi
}

# Show statistics
cmd_stats() {
    echo -e "${BLUE}Content Statistics:${NC}"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    
    local total=$(jq '.entries | length' "$DATA_FILE")
    local draft=$(jq '.entries | map(select(.status == "draft")) | length' "$DATA_FILE")
    local scheduled=$(jq '.entries | map(select(.status == "scheduled")) | length' "$DATA_FILE")
    local published=$(jq '.entries | map(select(.status == "published")) | length' "$DATA_FILE")
    
    echo -e "Total entries: ${GREEN}$total${NC}"
    echo -e "  Draft: ${YELLOW}$draft${NC}"
    echo -e "  Scheduled: ${BLUE}$scheduled${NC}"
    echo -e "  Published: ${GREEN}$published${NC}"
    
    echo -e "\nBy type:"
    jq -r '.entries | group_by(.type) | 
        map("  \(.[0].type): \(length)")[] ' "$DATA_FILE"
    
    echo -e "\nRecent activity:"
    jq -r '.entries | sort_by(.updated) | reverse | .[0:5] | 
        .[] | "  \(.updated | split("T")[0]) - \(.title)"' "$DATA_FILE"
}

# Main command dispatcher
case "${1:-}" in
    create) shift; cmd_create "$@" ;;
    list) shift; cmd_list "$@" ;;
    show) shift; cmd_show "$@" ;;
    update) shift; cmd_update "$@" ;;
    delete) shift; cmd_delete "$@" ;;
    schedule) shift; cmd_schedule "$@" ;;
    stats) shift; cmd_stats "$@" ;;
    -h|--help|help) print_usage ;;
    *)
        if [ -n "$1" ]; then
            echo -e "${RED}Unknown command: $1${NC}\n"
        fi
        print_usage
        exit 1
        ;;
esac