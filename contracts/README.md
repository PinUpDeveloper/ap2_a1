This folder simulates the **generated code repository** required by Assignment 2.

Expected flow:
1. Keep `.proto` files in `/protos` or in a dedicated proto repository.
2. Generate Go code into `contracts/gen/go/ap2proto`.
3. Import the generated package from both services.

Example generation commands (run locally or in CI):

```bash
protoc \
  -I ./protos \
  -I /usr/local/include \
  --go_out=./contracts/gen/go \
  --go-grpc_out=./contracts/gen/go \
  ./protos/payment.proto ./protos/order.proto
```

For full points, move `/protos` and `/contracts` into separate GitHub repositories and use remote generation.
