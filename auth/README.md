# auth

Reusable authentication primitives for HTTP services.  
Uses [github.com/golang-jwt/jwt/v5](https://github.com/golang-jwt/jwt) to sign and verify access/refresh tokens.  
Uses Go `context` values to propagate subject/admin flags through middleware and handlers.

## Usage

```bash
go get github.com/jesse0michael/pkg/auth
```
