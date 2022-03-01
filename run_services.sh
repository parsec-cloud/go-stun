#!/bin/bash

_term() {
  echo "Caught SIGTERM signal!"
  kill -TERM "$stun_service_pid" 2>/dev/null
  kill -TERM "$health_service_pid" 2>/dev/null
}

trap _term SIGTERM

# Start the stun server process
./stun-server/service &
stun_service_pid=$!

# Start the health check side-car service process
./health-server/service &
health_service_pid=$!

# Wait for any of the process to exit
wait -n
  
# Exit with status of process that exited first
exit $?
