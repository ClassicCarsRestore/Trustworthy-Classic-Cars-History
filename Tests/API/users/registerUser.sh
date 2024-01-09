#!/bin/bash

url="https://gui.classicschain.com:8393/api/Users/"
method="POST"

# Function to measure the time taken for an HTTP request with a JSON body
measure_request_time() {
   for ((i=0; i<=15; i++))
    do
        # Create the userID with the format "userX"
        userID="performance$i"

        # Create the JSON body with the userID
        # body="{
        #     \"email\": \"$userID\\@gmail.com\",
        #     \"password\": \"password123\",
        #     \"orgname\": \"Org1\"
        # }"

                body=$(cat <<EOF
{
    "email": "$userID@example.com",
    "password": "password123",
    "orgname": "Org1"
}
EOF
)

        start_time=$(date +%s.%N)
        response=$(curl -X "$method" -s -d "$body" -H "Content-Type: application/json" "$url")
        end_time=$(date +%s.%N)

        # Calculate the elapsed time
        elapsed_time=$(echo "$end_time - $start_time" | bc)

        # Print the result
        echo "HTTP $method request to $url returned response: $response"
        echo "Time taken: $elapsed_time seconds"
        echo "Email: $userID"
done
}

# Loop through the iterations


# Call the function to measure the request time
measure_request_time
