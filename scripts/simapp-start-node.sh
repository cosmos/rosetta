git clone https://github.com/cosmos/cosmos-sdk.git
# shellcheck disable=SC2164
cd cosmos-sdk
make build
export SIMD_BIN=./build/simd
chmod 777 ./scripts/init-simapp.sh
sh ./scripts/init-simapp.sh --just-print
$SIMD_BIN start &
until curl --output /dev/null --silent --head --fail http://localhost:26657/health; do
  sleep 1
done