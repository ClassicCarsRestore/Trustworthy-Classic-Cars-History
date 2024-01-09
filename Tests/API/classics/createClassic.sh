#!/bin/bash

url0="http://194.210.120.34:8393/api/Users/Login"
method="POST"
body0='{
    "email": "performanceOrg2@gmail.com",
    "password": "password123",
    "orgname": "Org2"
}'

url1="http://194.210.120.34:8393/api/Classics/Create"
body1='{
    "make": "Server",
    "model": "Server",
    "year": 1969,
    "licencePlate": "Server",
    "country": "Portugal",
    "chassisNo": "Server11",
    "engineNo": "Server",
    "ownerEmail": "murta2@gmail.com"
}'

# Function to measure the time taken for an HTTP request with a JSON body
measure_request_time() {
    response0=$(curl -X "$method" -s -d "$body0" -H "Content-Type: application/json" "$url0")

    userId=$(echo "$response0" | jq -r '.message.email')
    echo "User logged with email: $userId"

    auth_token=$(echo "$response0" | jq -r '.message.token')

    # Send the request and measure the time taken
    start_time=$(date +%s.%N)
    response1=$(curl -X "$method" -s -d "$body1" -H "Authorization: Bearer $auth_token" -H "Content-Type: application/json" "$url1")
    end_time=$(date +%s.%N)

    # Calculate the elapsed time
    elapsed_time=$(echo "$end_time - $start_time" | bc)
    formatted_elapsed_time=$(printf "%.7f" "$elapsed_time")
    formatted_elapsed_time=${formatted_elapsed_time/./,} # Replace period with comma


     # Print the result
    echo "HTTP $method request to $url1 returned response: $response1"
    echo "Time taken: $formatted_elapsed_time seconds"
}

# Call the function to measure the request time
measure_request_time
