#!/bin/bash

# TODO: remove sleep after lain has a more stable calico ip allocation

cd /lain/app

sleep 1

source ./secrets

EMAIL_DOMAIN=${EMAIL_DOMAIN:-"example.com"}
LAIN_DOMAIN=${LAIN_DOMAIN:-"example.com"}

DEBUG=${DEBUG:-"false"}
EMAIL_TLS=${EMAIL_TLS:-"false"}

if [ -n "$EMAIL_FROM" ];then
    FROM_USER_PASSWORD="$EMAIL_FROM"
fi

if [ -n "$EMAIL_FROM" ] && [ -n "$EMAIL_PASSWORD" ];then
    FROM_USER_PASSWORD="$EMAIL_FROM:$EMAIL_PASSWORD"
fi

if [ -z "$FROM_USER_PASSWORD" ]; then
    FROM_USER_PASSWORD="sso@example.com"
fi

exec ./sso-0.2.linux.amd64 -domain="@$EMAIL_DOMAIN" -from="$FROM_USER_PASSWORD"\
     -emailtls="$EMAIL_TLS" -mysql="$MYSQL" -site="https://sso.$LAIN_DOMAIN" -smtp="$SMTP" -web=":80" \
     -sentry="$SENTRY" -debug="$DEBUG"
