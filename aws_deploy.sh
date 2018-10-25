#!/bin/sh

aws cloudformation package --template-file template.yaml --s3-bucket fmt-api-bucket --output-template-file packaged-template.yaml

aws cloudformation deploy --template-file packaged-template.yaml --stack-name fmt-api-stack --capabilities CAPABILITY_IAM