url0="https://gui.classicschain.com:8393/api/Users/Login"
method="POST"
body0='{
    "email": "performance0@example.com",
    "password": "password123",
    "orgname": "Org1"
    
}'

url1="https://gui.classicschain.com:8393/api/Restorations/Get/performance0/performance0_step_"
method1="GET"

# Function to measure the time taken for an HTTP request with a JSON body
measure_request_time() {
    
    response0=$(curl -X "$method" -s -d "$body0" -H "Content-Type: application/json" "$url0")

    userId=$(echo "$response0" | jq -r '.message.email')
    echo "User logged with email: $userId"

    auth_token=$(echo "$response0" | jq -r '.message.token')

    for ((i=0; i<=15; i++))
    do 
        # Send the request and measure the time taken
        start_time=$(date +%s.%N)
        response1=$(curl -X "$method1" -s -d "$body1" -H "Authorization: Bearer $auth_token" "$url1$i")
        end_time=$(date +%s.%N)

        # Calculate the elapsed time
        elapsed_time=$(echo "$end_time - $start_time" | bc)
        formatted_elapsed_time=$(printf "%.7f" "$elapsed_time")
        formatted_elapsed_time=${formatted_elapsed_time/./,} # Replace period with comma

        # Print the result
        #echo "HTTP $method1 request to $url1 returned response: $response1"
        echo "HTTP $method1 request to $url1 returned"
        echo "Time taken: $formatted_elapsed_time seconds"
    done
}

# Call the function to measure the request time
measure_request_time