# IAM

[中文](./README-zh-CN.md)

## Install

### go install

```bash
go install github.com/erewhile/iam@latest
```

### Build from source

```bash
go install github.com/google/wire/cmd/wire@latest
git clone https://github.com/erewhile/iam.git
cd iam
wire ./internal/wire/
go generate ./internal/ent/generate.go
go build
# with debug
./iam.exe server --debug

# without debug
./iam.exe server
```

### Grab a prebuilt binary

Don't want to build it yourself? Just head over to the [Releases](https://github.com/erewhile/iam/releases) page and grab the binary for your platform.

## Configuration

```json
{
  "scheme": {
    "port": ":26621"
  },
  "database": {
    "host": "127.0.0.1",
    "port": "3306",
    "user": "root",
    "password": "root",
    "db_name": "iam",
    "timezone": "Asia/Shanghai",
    "max_idle_conns": 10,
    "max_open_conns": 100,
    "max_lifetime": 3600000000000
  },
  "token": {
    "kid": "iam-key-v1",
    "aad": "coKByVhWDMsgkMZwcQhKb2DPfSqQ2LFy",
    "access_token_ttl": 300000000000,
    "access_token_cookie_key": "atck",
    "refresh_token_ttl": 86400000000000,
    "refresh_token_cookie_key": "rtck"
  },
  "session": {
    "cookie_key": "iam_sid",
    "cookie_ttl": 28800000000000
  },
  "aes": {
    "key": "3Usycp2viFqS6RbBObq2JGsy2O3K6mjE"
  },
  "redis": {
    "addr": "127.0.0.1:6379",
    "password": "",
    "db": 0,
    "prefix": "iam",
    "pool_size": 100,
    "min_idle_conns": 10,
    "max_retries": 3,
    "dial_timeout": 5000000000,
    "read_timeout": 3000000000,
    "write_timeout": 3000000000,
    "pool_timeout": 4000000000
  },
  "logger": {
    "logs_dir": "logs",
    "max_size": 50,
    "max_backups": 10,
    "max_age": 24
  },
  "cors": {
    "allow_origins": [
      "http://127.0.0.1:26626",
      "http://localhost:26621"
    ],
    "allow_methods": [
      "GET",
      "POST",
      "PUT",
      "DELETE",
      "OPTIONS"
    ],
    "allow_headers": [
      "Origin",
      "Content-Type",
      "Accept",
      "Authorization"
    ],
    "allow_credentials": true,
    "max_age": 28800000000000
  },
  "login_security": {
    "MaxAttempts": 5,
    "AttemptWindow": 600000000000,
    "LockoutDuration": 900000000000,
    "MaxAttemptsByIP": 20
  },
  "hash": {
    "hmac_key": "NsefVBdLLOXtOVLPtBwFtIuB895ShuDw"
  }
}
```

What's what:

- **scheme** – the port the server listens on, e.g. `:26621`.
- **database** – your MySQL connection info. Time-ish fields like `max_lifetime` are in nanoseconds (`time.Duration`).
- **token** – signing keys for access/refresh tokens (`kid`, `aad`), their TTLs, and the cookie names they're stored under.
- **session** – cookie name and TTL for the session.
- **aes** – AES key used to encrypt/decrypt sensitive data.
- **redis** – Redis connection pool settings. `prefix` gets prepended to every key.
- **logger** – where logs go and how they rotate (max size per file, how many backups to keep, max age).
- **cors** – CORS settings. ⚠️ **`allow_origins` doesn't support `["*"]`** — you need to list out actual origins (scheme + host + port), like `http://localhost:26621`. Anything else gets blocked.
- **login_security** – brute-force protection: max failed attempts (per account / per IP), the time window for counting them, and how long an account gets locked out.
- **hash** – HMAC key used for signing/verifying.

> Heads up: anything that looks time-related (`*_ttl`, `*_timeout`, `max_lifetime`, `max_age`, etc.) is a `time.Duration` under the hood, so the raw JSON number is in nanoseconds.

## ⚖️ License

MIT License, see [LICENSE](./LICENSE) for details.