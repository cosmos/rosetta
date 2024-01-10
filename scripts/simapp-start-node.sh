git clone https://github.com/cosmos/cosmos-sdk.git
# shellcheck disable=SC2164
cd cosmos-sdk
make build
export SIMD_BIN=./build/simd
make init-simapp --dry-run
$SIMD_BIN start
until curl --output /dev/null --silent --head --fail http://localhost:26657/health; do
  echo "trying"
  sleep 1
done