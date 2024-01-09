#!/bin/bash

url0="http://194.210.120.34:8393/api/Users/Login"
method0="POST"
body0='{
    "email": "user1org1@gmail.com",
    "password": "password123",
    "orgname": "Org1"
}'

url01="http://194.210.120.34:8393/api/Restorations/Create/99997"
method01="POST"

url1="http://194.210.120.34:8393/api/Users/Login"
method1="POST"
body1='{
    "email": "raimundo.branco@gmail.com",
    "password": "password123",
    "orgname": "Org2"
}'

url12="http://194.210.120.34:8393/api/Restorations/Create/Pinterest"
method12="POST"

measure_request_time() {
    # Run the tasks in parallel using background processes
    start_time=$(date +%s.%N)
    (
        response0=$(curl -X "$method0" -s -d "$body0" -H "Content-Type: application/json" "$url0")
        userId0=$(echo "$response0" | jq -r '.message.email')
        echo "User logged with email: $userId0"
        auth_token0=$(echo "$response0" | jq -r '.message.token')
        #response2=$(curl -X "$method01" -s  -H "Authorization: Bearer $auth_token0" "$url01")
        response2=$(curl -X "$method01" -s -F "title=Primeiro titulo" -F "description=Primeira descricao" -H "Authorization: Bearer $auth_token0" -H "Content-Type: multipart/form-data" "$url01")
        echo "---------------------------RESPONSE22222------------------"
        echo "response2: $response2"
    ) &

    (
        response1=$(curl -X "$method1" -s -d "$body1" -H "Content-Type: application/json" "$url1")
        userId1=$(echo "$response1" | jq -r '.message.email')
        echo "User logged with email: $userId1"
        auth_token1=$(echo "$response1" | jq -r '.message.token')
        #response3=$(curl -X "$method12" -s  -H "Authorization: Bearer $auth_token1" "$url12")
        response3=$(curl -X "$method12" -s -F "title=segundo titulo" -F "description=segunda descricao" -H "Authorization: Bearer $auth_token1" -H "Content-Type: multipart/form-data" "$url12")
        echo "---------------------------RESPONSE33333------------------"
        echo "response3: $response3"
    ) &
    # Wait for all background processes to finish
    wait

    end_time=$(date +%s.%N)
    elapsed_time=$(echo "$end_time - $start_time" | bc)
    formatted_elapsed_time=$(printf "%.7f" "$elapsed_time")
    formatted_elapsed_time=${formatted_elapsed_time/./,} # Replace period with comma

    # Print the result
    echo "Time taken: $formatted_elapsed_time seconds"
}

# Call the function to measure the request time
measure_request_time
