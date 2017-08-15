#!/usr/bin/env bash

set -e
export GOPATH=/go
cd $GOPATH/src/github.com/laincloud/sso
echo "" > coverage.txt

for d in $(godep go list ./...); do
    TEMPLATES_PATH=$GOPATH/src/github.com/laincloud/sso/templates TEST_MYSQL_DSN="test:test@(127.0.0.1:3306)/sso_test" godep go test -v -p 1 -coverprofile=profile.out -covermode=atomic $d
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done
