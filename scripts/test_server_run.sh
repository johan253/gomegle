#!/bin/bash

# Run the server in the background
make run &
SERVER_PID=$!

# Wait a few seconds to let the server start
sleep 10

# Check if the process is still running
if ps -p $SERVER_PID >/dev/null; then
  echo "Server started successfully. Killing process..."
  kill $SERVER_PID
  wait $SERVER_PID 2>/dev/null
  echo "Process killed."
else
  echo "Server did not start."
  exit 1
fi
