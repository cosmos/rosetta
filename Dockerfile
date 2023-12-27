# ------------------------------------------------------------------------------
# Build gaia
# ------------------------------------------------------------------------------
FROM golang:1.20 AS cosmos

ARG COSMOS_VERSION

RUN git clone https://github.com/cosmos/cosmos-sdk \
  /go/src/github.com/cosmos/cosmos-sdk

WORKDIR /go/src/github.com/cosmos/cosmos-sdk

RUN git checkout v0.50.2 && make build
RUN export SIMD_BIN=/go/src/github.com/cosmos/cosmos-sdk/build/simd && make init-simapp

# ------------------------------------------------------------------------------
# Build rosetta
# ------------------------------------------------------------------------------
FROM golang:1.20.10 AS rosetta

ARG ROSETTA_VERSION

RUN git clone https://github.com/cosmos/rosetta.git \
  /go/src/github.com/cosmos/rosetta

WORKDIR /go/src/github.com/cosmos/rosetta

RUN git checkout $ROSETTA_VERSION && \
    go mod download

RUN make build


# ------------------------------------------------------------------------------
# Target container for running the node and rosetta server
# ------------------------------------------------------------------------------
FROM golang:1.20.10

# Install dependencies
RUN apt-get update -y && \
    apt-get install -y wget

WORKDIR /app

COPY --from=cosmos \
  /go/src/github.com/cosmos/cosmos-sdk/build/simd \
  /app/simd

# Install rosetta server
COPY --from=rosetta \
  /go/src/github.com/cosmos/rosetta/rosetta \
  /app/rosetta

## Install service start script
ADD scripts/entrypoint.sh /scripts/entrypoint.sh
USER root

RUN chmod +x /scripts/entrypoint.sh

EXPOSE 9650
EXPOSE 9651
EXPOSE 8080

ENTRYPOINT ["sh","/scripts/entrypoint.sh"]
STOPSIGNAL SIGTERM