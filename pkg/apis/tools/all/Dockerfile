FROM alpine:3.7 as prototool_download

RUN apk update && \
        apk add curl && \
        curl -sSL https://github.com/uber/prototool/releases/download/v1.3.0/prototool-Linux-x86_64 \
        -o /usr/bin/prototool && \
        chmod +x /usr/bin/prototool

FROM gcr.io/istio-operator/protolock:2019-04-10 as protolock

FROM gcr.io/istio-operator/protoc:2019-04-10

COPY --from=protolock \
        /usr/bin/protolock \
        /usr/bin/

COPY --from=prototool_download \
        /usr/bin/prototool \
        /usr/bin/

RUN apk update && \
        apk add libc6-compat

ENTRYPOINT []
