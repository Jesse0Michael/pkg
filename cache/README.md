# cache

Fault-tolerant caching primitives for Go services.  
Uses [go-redis/redis](https://github.com/go-redis/redis) for primary cache storage.  
Uses [go-redis/cache](https://github.com/go-redis/cache) TinyLFU to add local caching.  
Uses [sony/gobreaker](https://github.com/sony/gobreaker) to isolate Redis failures.  
Respects HTTP cache-control directives to opt-in/opt-out per request.

## Usage

```bash
go get github.com/jesse0michael/pkg/cache
```
