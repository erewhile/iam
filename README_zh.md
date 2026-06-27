# IAM

[English](./README.md)

## 安装

### go install

```bash
go install github.com/erewhile/iam@latest
```

### 源码安装

```bash
go install github.com/google/wire/cmd/wire@latest
git clone https://github.com/erewhile/iam.git
cd iam
wire ./internal/wire/
go generate ./internal/ent/generate.go
go build
# 调试模式
./iam.exe server --debug

# 正常模式
./iam.exe server
```

### 下载二进制文件

也可以直接前往 [Release](https://github.com/erewhile/iam/releases) 页面下载对应平台的预编译版本，无需自行编译。

## 配置文件

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

字段说明：

- **scheme**：服务监听端口，比如 `:26621`。
- **database**：MySQL 连接信息，`max_lifetime` 等时间字段单位是纳秒（`time.Duration`）。
- **token**：access token / refresh token 的签名密钥（`kid`、`aad`）、有效期（`ttl`）以及对应的 cookie 名称。
- **session**：会话 cookie 的名称和有效期。
- **aes**：用于敏感数据加解密的 AES 密钥。
- **redis**：Redis 连接池配置，`prefix` 是所有 key 的统一前缀。
- **logger**：日志文件目录及滚动策略（单文件大小、保留份数、保留天数）。
- **cors**：跨域配置。⚠️ **`allow_origins` 不支持 `["*"]`，必须填写具体的协议+域名/IP+端口**，例如 `http://localhost:26621`，否则跨域请求会被拒绝。
- **login_security**：登录失败保护策略，比如单账号/单 IP 的最大失败次数、统计窗口和锁定时长。
- **hash**：用于 HMAC 签名/校验的密钥。

> 时间相关的字段（`*_ttl`、`*_timeout`、`max_lifetime`、`max_age` 等）均为 `time.Duration`，对应的 JSON 数值单位是纳秒。

## ⚖️ 许可证

MIT License. 详见 [LICENSE](./LICENSE)。