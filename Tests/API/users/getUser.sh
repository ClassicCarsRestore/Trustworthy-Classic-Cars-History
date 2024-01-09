#!/bin/bash

url="https://gui.classicschain.com:8393/api/Users/Get/zmurta15@gmail.com"
method="GET"

# Function to measure the time taken for an HTTP request with a JSON body
measure_request_time() {
    # Send the request and measure the time taken
    for ((i=0; i<=15; i++))
        do
            start_time=$(date +%s.%N)
            response=$(curl -X "$method" -s "$url")
            end_time=$(date +%s.%N)

            # Calculate the elapsed time
            elapsed_time=$(echo "$end_time - $start_time" | bc)

            # Print the result
            echo "HTTP $method request to $url returned response: $response"
            #echo "HTTP $method request to $url returned"
            echo "Time taken: $elapsed_time seconds"
    done
}

# Call the function to measure the request time
measure_request_time
