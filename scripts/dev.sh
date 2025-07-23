#!/bin/bash

# Development script for GoMegle
# Watches for file changes and automatically restarts the application

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[DEV]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[DEV]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[DEV]${NC} $1"
}

print_error() {
    echo -e "${RED}[DEV]${NC} $1"
}

# Check if inotifywait is installed
if ! command -v inotifywait &> /dev/null; then
    print_error "inotifywait is not installed. Please install inotify-tools:"
    print_error "  Ubuntu/Debian: sudo apt-get install inotify-tools"
    print_error "  Fedora/RHEL:   sudo dnf install inotify-tools"
    print_error "  Arch:          sudo pacman -S inotify-tools"
    exit 1
fi

# PID of the currently running process
APP_PID=""

# Function to kill the running process
kill_process() {
    if [[ -n "$APP_PID" ]] && kill -0 "$APP_PID" 2>/dev/null; then
        print_warning "Stopping running process (PID: $APP_PID)..."
        kill "$APP_PID" 2>/dev/null || true
        wait "$APP_PID" 2>/dev/null || true
        APP_PID=""
    fi
}

# Function to start the application
start_app() {
    print_status "Building and starting application..."
    make run &
    APP_PID=$!
    print_success "Application started (PID: $APP_PID)"
}

# Function to restart the application
restart_app() {
    print_status "File change detected, restarting..."
    kill_process
    sleep 0.5  # Brief pause to ensure cleanup
    start_app
}

# Cleanup function for graceful exit
cleanup() {
    print_warning "Shutting down development server..."
    kill_process
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

print_status "Starting GoMegle development server..."
print_status "Watching for changes in: *.go files"
print_status "Press Ctrl+C to stop"

# Start the application initially
start_app

# Watch for file changes
while true; do
    # Watch for modifications to .go files in current directory and subdirectories
    inotifywait -e modify,create,delete,move \
        --include '.*\.go$' \
        -r . \
        --quiet 2>/dev/null || {
        # If inotifywait exits unexpectedly, restart it
        print_warning "File watcher stopped unexpectedly, restarting..."
        sleep 1
        continue
    }
    
    # Restart the application when changes are detected
    restart_app
    
    # Brief cooldown to avoid rapid restarts
    sleep 1
done
