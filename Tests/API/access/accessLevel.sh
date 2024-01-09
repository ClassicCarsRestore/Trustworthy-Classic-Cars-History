url0="http://194.210.120.34:8393/api/Users/Login"
method="POST"
body0='{
    "email": "murta2@gmail.com",
    "password": "password123",
    "orgname": "Org1"
    
}'

url1="http://194.210.120.34:8393/api/Access/CheckLevel/Server2"
method1="GET"

measure_request_time() {


    for ((i=0; i<=15; i++))
    do 
        response0=$(curl -X "$method" -s -d "$body0" -H "Content-Type: application/json" "$url0")

        userId=$(echo "$response0" | jq -r '.message.email')
        echo "User logged with email: $userId"

        auth_token=$(echo "$response0" | jq -r '.message.token')
        # Send the request and measure the time taken
        start_time=$(date +%s.%N)
        response1=$(curl -X "$method1" -s  -H "Authorization: Bearer $auth_token" "$url1")
        end_time=$(date +%s.%N)

        # Calculate the elapsed time
        elapsed_time=$(echo "$end_time - $start_time" | bc)
        formatted_elapsed_time=$(printf "%.7f" "$elapsed_time")
        formatted_elapsed_time=${formatted_elapsed_time/./,} # Replace period with comma

        # Print the result
        echo "HTTP $method1 request to $url1 returned response: $response1"
        #echo "HTTP $method1 request to $url1 returned"
        echo "Time taken: $formatted_elapsed_time seconds"

    done
}

# Call the function to measure the request time
measure_request_time