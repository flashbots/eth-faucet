# stage: build ---------------------------------------------------------

FROM golang:1.22-alpine as build

RUN apk add --no-cache gcc musl-dev linux-headers

WORKDIR /go/src/github.com/flashbots/eth-faucet

COPY go.* ./
RUN go mod download

COPY . .

RUN go build -o bin/eth-faucet -ldflags "-s -w" github.com/flashbots/eth-faucet/cmd

# stage: run -----------------------------------------------------------

FROM alpine

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=build /go/src/github.com/flashbots/eth-faucet/bin/eth-faucet ./eth-faucet

ENTRYPOINT ["/app/eth-faucet"]
