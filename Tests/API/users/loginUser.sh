#!/bin/bash

url="https://gui.classicschain.com:8393/api/Users/Login"
method="POST"
body='{
    "email": "performance0@example.com",
    "password": "password123",
    "orgname": "Org1"
    
}'

# Function to measure the time taken for an HTTP request with a JSON body
measure_request_time() {
    # Send the request and measure the time taken
    for ((i=0; i<=15; i++))
        do
            start_time=$(date +%s.%N)
            response=$(curl -X "$method" -s -d "$body" -H "Content-Type: application/json" "$url")
            end_time=$(date +%s.%N)

            # Calculate the elapsed time
            elapsed_time=$(echo "$end_time - $start_time" | bc)

            # Print the result
            echo "HTTP $method request to $url returned response: $response"
            echo "Time taken: $elapsed_time seconds"

            userId=$(echo "$response" | jq -r '.message.email')
            echo "Email: $userId"~
    done
}

# Call the function to measure the request time
measure_request_time
