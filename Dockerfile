FROM golang:1.20 AS build-env

# Set working directory for the build
WORKDIR /go/src/github.com/cosmos/rosetta

# optimization: if go.sum didn't change, docker will use cached image
COPY go.mod go.sum ./

RUN go mod download

# Add source files
COPY . .

RUN make build
RUN make plugin

EXPOSE 8080

# Run simd by default
CMD ["./rosetta", "--blockchain", "app", "--network", "network", "--tendermint", "cosmos:26657", "--grpc", "cosmos:9090", "--addr", ":8080"]
ENTRYPOINT "./rosetta"
STOPSIGNAL SIGTERM