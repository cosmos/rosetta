FROM golang:1.20-alpine AS build-env

# Install minimum necessary dependencies
ENV PACKAGES curl make
RUN apk add --no-cache $PACKAGES

# Set working directory for the build
WORKDIR /go/src/github.com/cosmos/rosetta

# optimization: if go.sum didn't change, docker will use cached image
COPY go.mod go.sum ./

RUN go mod download

# Add source files
COPY . .

RUN make build

EXPOSE 8080

# Run simd by default
CMD ["./rosetta"]
STOPSIGNAL SIGTERM