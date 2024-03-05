# eth-faucet

Inspired and significantly influenced by https://github.com/chainflag/eth-faucet

## TL;DR

- Run

  ```shell
  make docker-compose
  ```

- And then open https://localhost:5173

## Features

- Authentication with twitter or github.
- Rate-limiting with redis.

## Configuration

>
> [!IMPORTANT]
>
> Backend and frontend must share the same `AUTH_SECRET` so that JWT api tokens
> could be validated.
>

### Backend configuration

Both, environment variables and command line switches are possible to use.

```text
CHAIN:

--chain-name name            chain name (default: "testnet") [$FAUCET_CHAIN_NAME]
--chain-token-symbol symbol  token symbol (default: "tEth") [$FAUCET_CHAIN_TOKEN_SYMBOL]

FAUCET:

--faucet-interval duration                       minimum duration to wait between funding rounds (default: 15m0s) [$FAUCET_INTERVAL]
--faucet-interval-address duration               minimum duration to wait between funding rounds for the same receiving address (default: 15m0s) [$FAUCET_INTERVAL_ADDRESS]
--faucet-interval-identity duration              minimum duration to wait between funding rounds for the same identity (default: 15m0s) [$FAUCET_INTERVAL_IDENTITY]
--faucet-interval-identity-and-address duration  minimum duration to wait between funding rounds for the same identity and receiving address (default: 15m0s) [$FAUCET_INTERVAL_IDENTITY_AND_ADDRESS]
--faucet-interval-ip duration                    minimum duration to wait between funding rounds for the same source IP (default: 15m0s) [$FAUCET_INTERVAL_IP]
--faucet-payout number                           number of tokens to transfer per user request (default: 1) [$FAUCET_PAYOUT]

REDIS:

--redis-namespace namespace  rate-limiting redis namespace (default: "eth-faucet") [$FAUCET_REDIS_NAMESPACE]
--redis-timeout timeout      timeout for redis operations (default: 200ms) [$FAUCET_REDIS_TIMEOUT]
--redis-url url              redis url for rate-limiting (default: "redis://localhost:6379") [$FAUCET_REDIS_URL]

RPC:

--rpc-endpoint endpoint  endpoint for ethereum json-rpc connection (default: "http://localhost:8545") [$FAUCET_RPC_ENDPOINT]
--rpc-timeout timeout    timeout for ethereum json-rpc operations (default: 5s) [$FAUCET_RPC_TIMEOUT]

SERVER:

--server-auth-secret secret           jwt authentication secret [$FAUCET_SERVER_AUTH_SECRET, $AUTH_SECRET]
--server-listen-address host:port     host:port for the server to listen on (default: "0.0.0.0:8080") [$FAUCET_SERVER_LISTEN_ADDRESS]
--server-max-request-body-size bytes  max request body size in bytes (default: 1024) [$FAUCET_SERVER_MAX_REQUEST_BODY_SIZE]
--server-proxy-count count            count of reverse proxies in front of the server (default: 0) [$FAUCET_SERVER_PROXY_COUNT]

WALLET:

--wallet-keystore json-file          funding wallet's keystore json-file [$FAUCET_WALLET_KEYSTORE]
--wallet-keystore-password password  funding wallet's keystore password [$FAUCET_WALLET_KEYSTORE_PASSWORD]
--wallet-private-key hex             funding wallet's private key hex [$FAUCET_WALLET_PRIVATE_KEY]
```

### Frontend configuration

Frontend is configured with environment variables (or with [`dotfiles`](https://www.npmjs.com/package/dotfiles)).

```shell
# 32 random hexadecimals (e.g. openssl rand -hex 32)
AUTH_SECRET="0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

# oauth 2.0 client id
AUTH_TWITTER_ID="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
# oauth 2.0 client secret
AUTH_TWITTER_SECRET="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# github app id
AUTH_GITHUB_ID="xxxxxxxxxxxxxxxxxxxx"
# github app secret
AUTH_GITHUB_SECRET="xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"

# base host from which the faucet is serving
ORIGIN=https://f.q.d.n.com
```
