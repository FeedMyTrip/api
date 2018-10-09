#!/bin/sh
cd functions

GOOS=linux
export GOOS

GOARCH=amd64
export GOARCH

for d in * ; do
    echo "Building function: $d"

    go build -o ./$d/fmt-lambda-$d ./$d/$d.go

    build-lambda-zip -o ../deploy/$d.zip ./$d/fmt-lambda-$d

    rm ./$d/fmt-lambda-$d

done

echo "Build finished finished"

cd ../

aws cloudformation package --template-file template.yaml --s3-bucket fmt-api-bucket --output-template-file packaged-template.yaml

aws cloudformation deploy --template-file packaged-template.yaml --stack-name fmt-api-stack --capabilities CAPABILITY_IAM

echo "FMT AWS deploy finished"