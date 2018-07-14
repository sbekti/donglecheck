FROM golang:alpine AS build-env
RUN apk add --update tzdata bash wget curl git
RUN mkdir -p $$GOPATH/bin && \
    curl https://glide.sh/get | sh
ADD . /go/src/donglecheck
WORKDIR /go/src/donglecheck
RUN glide update && go build -o main

FROM alpine
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build-env /go/src/donglecheck/main /app/
ENTRYPOINT ["./main"]
