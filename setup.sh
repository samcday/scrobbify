#!/bin/bash

set -uex -o pipefail

export AWS_ACCESS_KEY_ID=foo AWS_SECRET_ACCESS_KEY=foo

aws dynamodb --endpoint=http://localhost:8000 create-table --table-name users --attribute-definitions AttributeName=id,AttributeType=S --key-schema AttributeName=id,KeyType=HASH --provisioned-throughput ReadCapacityUnits=5,WriteCapacityUnits=5
