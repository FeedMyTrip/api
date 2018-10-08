set GOOS=linux
set GOARCH=amd64
go build -o fmt-lambda-trips
build-lambda-zip -o ../deploy/trips.zip fmt-lambda-trips
rm fmt-lambda-trips