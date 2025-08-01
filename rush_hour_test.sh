#!/bin/bash
# rush_hour.sh
# This script simulates 15 elevator requests during rush hour.
# There are three elevators:
#   - Elevator 1 and Elevator 2 serve floors 0 to 9.
#   - Elevator 3 serves floors from -4 (parking) to 5.
# The script generates a random valid request for each elevator,
# sends a curl POST request to the API endpoint,
# and sleeps for a random interval between requests.

# Function to generate a random number between two values (inclusive)
generate_random() {
    local min=$1
local max=$2
echo $(( RANDOM % (max - min + 1) + min ))
}

# Total number of simulated requests
total_requests=30

# Loop for each simulated request
for (( i=1; i<=total_requests; i++ )); do
# Randomly select an elevator type: 1 or 2 for two elevators, 3 for the special one.
elevator=$(( RANDOM % 3 + 1 ))

# Initialize floor range variables
if [[ $elevator -eq 3 ]]; then
# Elevator 3: floors from -4 to 5
min_floor=-4
max_floor=5
else
# Elevator 1 and 2: floors from 0 to 9
min_floor=0
max_floor=9
fi

# Generate a valid "from" floor
from=$(generate_random $min_floor $max_floor)

# Generate a valid "to" floor ensuring it is different from "from"
while true; do
to=$(generate_random $min_floor $max_floor)
if [[ $to -ne $from ]]; then
break
fi
done

# Print which elevator and the floors for logging
echo "Request $i: Elevator $elevator - from floor $from to floor $to"

# Send the POST request with JSON payload using curl
curl -X POST \
    -H "Content-Type: application/json" \
    -d "{\"from\": $from, \"to\": $to}" \
    http://localhost:6660/floor

echo -e "\n"  # Print newline for separation

# Sleep for a random time between 1 and 3 seconds to simulate real-life intervals
sleep_interval=$(generate_random 1 3)
sleep $sleep_interval
done
