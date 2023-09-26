#!/bin/bash

SIMD_BIN=${SIMD_BIN:=$(which simd 2>/dev/null)}

if [ -z "$SIMD_BIN" ]; then echo "SIMD_BIN is not set. Make sure to run make install before"; exit 1; fi

# Global variables
KEY_RING_BACKEND="test"
CHAIN_ID="demo"
amount_of_wallets=100

validator_address=$($SIMD_BIN keys show alice -a --keyring-backend "$KEY_RING_BACKEND")

echo "[INFO] Generating wallets: $amount_of_wallets"
for ((index=0; index<"$amount_of_wallets"; index++))
do
  random_number=$((1 + RANDOM % 10000))

  wallet_name="wallet-$random_number"
  $SIMD_BIN keys add $wallet_name --keyring-backend "$KEY_RING_BACKEND"
  wallet_address=$($SIMD_BIN keys show $wallet_name -a --keyring-backend "$KEY_RING_BACKEND")
  echo "[DEBUG] Generated wallet: $wallet_name - $wallet_address"

  amount="$((RANDOM*10))stake"
  echo "[DEBUG] Generating tx from validator: $validator_address sending $amount to $wallet_address"
  $SIMD_BIN tx bank send "$validator_address" "$wallet_address" "$amount" --chain-id "$CHAIN_ID"  --keyring-backend "$KEY_RING_BACKEND" -y

  sleep 1s # Wait so the TXs can happen in different blocks
done
