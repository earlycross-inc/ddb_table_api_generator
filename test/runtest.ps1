# start DynamoDB Local on docker
$container_name = docker run -d -p "8100:8000" amazon/dynamodb-local -jar DynamoDBLocal.jar -sharedDb

# generate table API
Push-Location ..
packr2.exe

go build -o gen.exe
.\gen.exe -def .\test\tbldef.test.yaml -out .\test\ddbtbl
Remove-Item .\gen.exe

packr2.exe clean
Pop-Location

# run tests
go test .

# clean up docker container
docker stop $container_name
docker rm $container_name


