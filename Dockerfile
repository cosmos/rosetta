# ------------------------------------------------------------------------------
# Build gaia
# ------------------------------------------------------------------------------
FROM golang:1.22 AS cosmos

ARG COSMOS_VERSION

RUN git clone https://github.com/cosmos/cosmos-sdk \
  /go/src/github.com/cosmos/cosmos-sdk

WORKDIR /go/src/github.com/cosmos/cosmos-sdk

RUN git checkout v0.50.7 && make build
RUN export SIMD_BIN=/go/src/github.com/cosmos/cosmos-sdk/build/simd && make init-simapp

# ------------------------------------------------------------------------------
# Build rosetta
# ------------------------------------------------------------------------------
FROM golang:1.22 AS rosetta

ARG ROSETTA_VERSION

RUN git clone https://github.com/cosmos/rosetta.git \
  /go/src/github.com/cosmos/rosetta

WORKDIR /go/src/github.com/cosmos/rosetta

RUN git checkout $ROSETTA_VERSION && \
    go mod download

RUN make build
RUN cd plugins/cosmos-hub && make plugin

COPY --from=cosmos \
  /go/src/github.com/cosmos/cosmos-sdk/build/simd \
  /app/simd

COPY --from=cosmos \
    /root/.simapp \
    /root/.simapp

ADD scripts/entrypoint.sh /scripts/entrypoint.sh
RUN chmod +x /scripts/entrypoint.sh

EXPOSE 9650
EXPOSE 9651
EXPOSE 8080

ENTRYPOINT ["sh","/scripts/entrypoint.sh"]
STOPSIGNAL SIGTERM