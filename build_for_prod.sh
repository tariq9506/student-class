#!/bin/bash
rm student-apis student-apis.tar.gz
/usr/local/go/bin/go get
GOOS=linux GOARCH=amd64 /usr/local/go/bin/go build -o ./student-apis
#upx student-apis
tar -cvzf student-apis.tar.gz database/ makefile student-apis .env.sample
