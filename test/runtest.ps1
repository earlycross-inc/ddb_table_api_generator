# start DynamoDB Local on docker
$container_name = docker run -d -p "8100:8000" amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb

# generate table API
pushd ..
go run .\... -def .\test\tbldef.test.yaml -out .\test\ddbtbl
popd

# run tests
go test .

# clean up docker container
docker stop $container_name
docker rm $container_name