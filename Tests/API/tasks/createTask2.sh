url0="https://gui.classicschain.com:8393/api/Users/Login"
method="POST"
body0='{
    "email": "murta2@gmail.com",
    "password": "password123",
    "orgname": "Org1"
    
}'

url1="https://gui.classicschain.com:8393/api/Restorations/Create/Server6"
method1="POST"

file_path0="../testDocs/download.jpg"
file_path1="../testDocs/pneus.jpg"
file_path2="../testDocs/rolls0.jpg"
file_path3="../testDocs/rolls1.jpg"
file_path4="../testDocs/rolls2.jpg"

measure_request_time() {


    for ((i=0; i<=15; i++))
    do 
        title="$i Collect vehicle information (schemes, data and paint)"
        description="$i Paint system and technic according to the original"
        response0=$(curl -X "$method" -s -d "$body0" -H "Content-Type: application/json" "$url0")

        userId=$(echo "$response0" | jq -r '.message.email')
        echo "User logged with email: $userId"

        auth_token=$(echo "$response0" | jq -r '.message.token')
        # Send the request and measure the time taken
        start_time=$(date +%s.%N)
        #response1=$(curl -X "$method1" -s -F "name=$fileName" -F "file=@$file_path" -H "Authorization: Bearer $auth_token" -H "Content-Type: multipart/form-data" "$url1")
        response1=$(curl -X "$method1" -s -F "title=$title" -F "description=$description" -F "file=@$file_path0" -F "file=@$file_path1" -F "file=@$file_path2" -F "file=@$file_path3" -F "file=@$file_path4" -H "Authorization: Bearer $auth_token" -H "Content-Type: multipart/form-data" "$url1")
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