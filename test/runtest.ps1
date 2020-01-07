# start DynamoDB Local on docker
$container_name = docker run -d -p "8100:8000" amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb

# generate table API
pushd ..
go build -o gen.exe
.\gen.exe -def .\test\tbldef.test.yaml -out .\test\ddbtbl
rm .\gen.exe
popd

# run tests
go test .

# clean up docker container
docker stop $container_name
docker rm $container_name