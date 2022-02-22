#!/bin/bash

# Start the stun server process
./stun-server/service &
  
# Start the health check side-car service process
./health-server/service &
  
# Wait for any of the process to exit
wait -n
  
# Exit with status of process that exited first
exit $?