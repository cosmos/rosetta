git clone https://github.com/cosmos/cosmos-sdk.git
# shellcheck disable=SC2164
cd cosmos-sdk
make build
export SIMD_BIN=./build/simd
make init-simapp --dry-run
echo "simapp started"
$SIMD_BIN start &
echo "start"
until curl --silent --head --fail http://localhost:26657/health; do
  sleep 1
done