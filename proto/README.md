# pkg Proto Options

Reusable Protobuf custom options for gRPC services.

## Options

### Method Options

Apply these to individual RPCs in your `.proto` files:

| Option | Field Number | Description |
|---|---|---|
| `no_auth` | 50000 | Bypass authentication for this RPC |
| `admin_only` | 50001 | Restrict access to admin users |
| `reject_read_only` | 50002 | Reject read-only users |

### Service Options

Apply these to an entire service:

| Option | Field Number | Description |
|---|---|---|
| `service_no_auth` | 50000 | Bypass authentication for all RPCs in the service |
| `service_admin_only` | 50001 | Restrict all RPCs in the service to admin users |

## Usage

```protobuf
import "options/v1/auth.proto";

service MyService {
  rpc PublicEndpoint(Request) returns (Response) {
    option (options.v1.no_auth) = true;
  }

  rpc AdminEndpoint(Request) returns (Response) {
    option (options.v1.admin_only) = true;
  }

  rpc WriteEndpoint(Request) returns (Response) {
    option (options.v1.reject_read_only) = true;
  }
}

service InternalService {
  option (options.v1.service_admin_only) = true;

  rpc DoWork(Request) returns (Response) {}
}
```
