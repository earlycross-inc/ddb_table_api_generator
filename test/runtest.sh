#!/bin/bash

# Start DynamoDB Local on Docker
container_name=$(docker run -d --rm -p "8100:8000" amazon/dynamodb-local:latest -jar DynamoDBLocal.jar -sharedDb)

# Generate table API
pushd ..

go build -o gen
./gen -def ./test/tbldef.test.yaml -api -api_out ./test/ddbtbl
rm ./gen

popd

# Run tests
go test .

# Clean up Docker container
docker stop "$container_name"
