#!/bin/bash

# TODO: remove sleep after lain has a more stable calico ip allocation

sleep 5

DOMAIN=${LAIN_DOMAIN:-"example.com"}
source ./secrets

DEBUG=${DEBUG:-"false"}

exec ./sso-0.2.linux.amd64 -domain="@$EMAIL" -from="sso@$DOMAIN" -mysql="$MYSQL" -site="https://sso.$DOMAIN" -smtp="$SMTP" -web=":80" -sentry="$SENTRY" -debug="$DEBUG"
