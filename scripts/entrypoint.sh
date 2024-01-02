/app/simd start &
until curl --output /dev/null --silent --head --fail http://localhost:26657/health; do
   sleep 1
done
./rosetta --blockchain "cosmos" --network "cosmos" --tendermint "tcp://localhost:26657" --addr "localhost:8080" --grpc "localhost:9090"