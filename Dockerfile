FROM golang:1.21-alpine3.18 as builder
ARG IBC_GO_VERSION

RUN set -eux; apk add --no-cache git libusb-dev linux-headers gcc musl-dev make;

ENV GOPATH=""
ENV IBC_GO_VERSION=v8.0.0-beta.1

# ensure the ibc go version is being specified for this image.
RUN test -n "${IBC_GO_VERSION}"

# Copy relevant files before go mod download. Replace directives to local paths break if local
# files are not copied before go mod download.
ADD abci abci
ADD api api
ADD keeper keeper
ADD migrations migrations
ADD module module
ADD simapp simapp

COPY go.mod .
COPY go.sum .

# Copy all .go files from current directory to the Docker image
COPY *.go ./

RUN go mod download

RUN GOOS=linux GOARCH=amd64 LEDGER_ENABLED=false go build -mod=readonly -tags "netgo ledger" -ldflags '-X github.com/cosmos/cosmos-sdk/version.Name=sim -X github.com/cosmos/cosmos-sdk/version.AppName=simd -X github.com/cosmos/cosmos-sdk/version.Version= -X github.com/cosmos/cosmos-sdk/version.Commit= -X "github.com/cosmos/cosmos-sdk/version.BuildTags=netgo ledger," -w -s' -trimpath -o /go/build/ ./...

FROM alpine:3.18
ARG IBC_GO_VERSION

LABEL "org.cosmos.ibc-go" "${IBC_GO_VERSION}"

COPY --from=builder /go/build/simd /bin/simd

ENTRYPOINT ["simd"]
