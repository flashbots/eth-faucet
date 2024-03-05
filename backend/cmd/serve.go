package main

import (
	"slices"
	"time"

	"github.com/flashbots/eth-faucet/config"
	"github.com/flashbots/eth-faucet/server"
	"github.com/urfave/cli/v2"
)

const (
	categoryChain  = "CHAIN:"
	categoryFaucet = "FAUCET:"
	categoryRedis  = "REDIS:"
	categoryRPC    = "RPC:"
	categoryServer = "SERVER:"
	categoryWallet = "WALLET:"
)

func CommandServe(cfg *config.Config) *cli.Command {
	chainFlags := []cli.Flag{
		&cli.StringFlag{
			Category:    categoryChain,
			Destination: &cfg.Chain.Name,
			EnvVars:     []string{"FAUCET_CHAIN_NAME"},
			Name:        "chain-name",
			Usage:       "chain `name`",
			Value:       "testnet",
		},

		&cli.StringFlag{
			Category:    categoryChain,
			Destination: &cfg.Chain.TokenSymbol,
			EnvVars:     []string{"FAUCET_CHAIN_TOKEN_SYMBOL"},
			Name:        "chain-token-symbol",
			Usage:       "token `symbol`",
			Value:       "tEth",
		},
	}

	faucetFlags := []cli.Flag{
		&cli.DurationFlag{
			Category:    categoryFaucet,
			Destination: &cfg.Faucet.Interval,
			EnvVars:     []string{"FAUCET_INTERVAL"},
			Name:        "faucet-interval",
			Usage:       "minimum `duration` to wait between funding rounds",
			Value:       15 * time.Minute,
		},

		&cli.DurationFlag{
			Category:    categoryFaucet,
			Destination: &cfg.Faucet.IntervalAddress,
			EnvVars:     []string{"FAUCET_INTERVAL_ADDRESS"},
			Name:        "faucet-interval-address",
			Usage:       "minimum `duration` to wait between funding rounds for the same receiving address",
			Value:       15 * time.Minute,
		},

		&cli.DurationFlag{
			Category:    categoryFaucet,
			Destination: &cfg.Faucet.IntervalIdentity,
			EnvVars:     []string{"FAUCET_INTERVAL_IDENTITY"},
			Name:        "faucet-interval-identity",
			Usage:       "minimum `duration` to wait between funding rounds for the same identity",
			Value:       15 * time.Minute,
		},

		&cli.DurationFlag{
			Category:    categoryFaucet,
			Destination: &cfg.Faucet.IntervalIdentityAndAddress,
			EnvVars:     []string{"FAUCET_INTERVAL_IDENTITY_AND_ADDRESS"},
			Name:        "faucet-interval-identity-and-address",
			Usage:       "minimum `duration` to wait between funding rounds for the same identity and receiving address",
			Value:       15 * time.Minute,
		},

		&cli.DurationFlag{
			Category:    categoryFaucet,
			Destination: &cfg.Faucet.IntervalIP,
			EnvVars:     []string{"FAUCET_INTERVAL_IP"},
			Name:        "faucet-interval-ip",
			Usage:       "minimum `duration` to wait between funding rounds for the same source IP",
			Value:       15 * time.Minute,
		},

		&cli.Int64Flag{
			Category:    categoryFaucet,
			Destination: &cfg.Faucet.Payout,
			EnvVars:     []string{"FAUCET_PAYOUT"},
			Name:        "faucet-payout",
			Usage:       "`number` of tokens to transfer per user request",
			Value:       1,
		},
	}

	redisFlags := []cli.Flag{
		&cli.StringFlag{
			Category:    categoryRedis,
			Destination: &cfg.Redis.Namespace,
			EnvVars:     []string{"FAUCET_REDIS_NAMESPACE"},
			Name:        "redis-namespace",
			Usage:       "rate-limiting redis `namespace`",
			Value:       "eth-faucet",
		},

		&cli.DurationFlag{
			Category:    categoryRedis,
			Destination: &cfg.Redis.Timeout,
			EnvVars:     []string{"FAUCET_REDIS_TIMEOUT"},
			Name:        "redis-timeout",
			Usage:       "`timeout` for redis operations",
			Value:       200 * time.Millisecond,
		},

		&cli.StringFlag{
			Category:    categoryRedis,
			Destination: &cfg.Redis.URL,
			EnvVars:     []string{"FAUCET_REDIS_URL"},
			Name:        "redis-url",
			Usage:       "redis `url` for rate-limiting",
			Value:       "redis://localhost:6379",
		},
	}

	rpcFlags := []cli.Flag{
		&cli.StringFlag{
			Category:    categoryRPC,
			Destination: &cfg.RPC.Endpoint,
			EnvVars:     []string{"FAUCET_RPC_ENDPOINT"},
			Name:        "rpc-endpoint",
			Usage:       "`endpoint` for ethereum json-rpc connection",
			Value:       "http://localhost:8545",
		},

		&cli.DurationFlag{
			Category:    categoryRPC,
			Destination: &cfg.RPC.Timeout,
			EnvVars:     []string{"FAUCET_RPC_TIMEOUT"},
			Name:        "rpc-timeout",
			Usage:       "`timeout` for ethereum json-rpc operations",
			Value:       5 * time.Second,
		},
	}

	serverFlags := []cli.Flag{
		&cli.StringFlag{
			Category:    categoryServer,
			Destination: &cfg.Server.AuthSecret,
			EnvVars:     []string{"FAUCET_SERVER_AUTH_SECRET", "AUTH_SECRET"},
			Name:        "server-auth-secret",
			Usage:       "jwt authentication `secret`",
		},

		&cli.StringFlag{
			Category:    categoryServer,
			Destination: &cfg.Server.ListenAddress,
			EnvVars:     []string{"FAUCET_SERVER_LISTEN_ADDRESS"},
			Name:        "server-listen-address",
			Usage:       "`host:port` for the server to listen on",
			Value:       "0.0.0.0:8080",
		},

		&cli.IntFlag{
			Category:    categoryServer,
			Destination: &cfg.Server.MaxRequestBodySize,
			EnvVars:     []string{"FAUCET_SERVER_MAX_REQUEST_BODY_SIZE"},
			Name:        "server-max-request-body-size",
			Usage:       "max request body size in `bytes`",
			Value:       1024,
		},

		&cli.IntFlag{
			Category:    categoryServer,
			Destination: &cfg.Server.ProxyCount,
			EnvVars:     []string{"FAUCET_SERVER_PROXY_COUNT"},
			Name:        "server-proxy-count",
			Usage:       "`count` of reverse proxies in front of the server",
			Value:       0,
		},
	}

	walletFlags := []cli.Flag{
		&cli.StringFlag{
			Category:    categoryWallet,
			Destination: &cfg.Wallet.Keystore,
			EnvVars:     []string{"FAUCET_WALLET_KEYSTORE"},
			Name:        "wallet-keystore",
			Usage:       "funding wallet's keystore `json-file`",
		},

		&cli.StringFlag{
			Category:    categoryWallet,
			Destination: &cfg.Wallet.KeystorePassword,
			EnvVars:     []string{"FAUCET_WALLET_KEYSTORE_PASSWORD"},
			Name:        "wallet-keystore-password",
			Usage:       "funding wallet's keystore `password`",
		},

		&cli.StringFlag{
			Category:    categoryWallet,
			Destination: &cfg.Wallet.PrivateKey,
			EnvVars:     []string{"FAUCET_WALLET_PRIVATE_KEY"},
			Name:        "wallet-private-key",
			Usage:       "funding wallet's private key `hex`",
		},
	}

	flags := slices.Concat(
		chainFlags,
		faucetFlags,
		redisFlags,
		rpcFlags,
		serverFlags,
		walletFlags,
	)

	return &cli.Command{
		Name:  "serve",
		Usage: "run the api server",
		Flags: flags,

		Action: func(_ *cli.Context) error {
			s, err := server.New(cfg)
			if err != nil {
				return err
			}
			return s.Run()
		},
	}
}
