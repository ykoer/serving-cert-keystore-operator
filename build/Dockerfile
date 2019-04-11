FROM alpine:3.8

RUN apk upgrade --update --no-cache

USER nobody

ADD build/_output/bin/serving-cert-keystore-operator /usr/local/bin/serving-cert-keystore-operator
