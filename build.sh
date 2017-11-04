#!/bin/bash

# Deploy Minio
helm install --set accessKey=myaccesskey,secretKey=mysecretkey,persistence.size=1Gi \
	stable/minio --name minio --version 0.3.2

# Build storage image
docker build -t local/storage ./storage
