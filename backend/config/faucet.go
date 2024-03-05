package config

import (
	"math/big"
	"time"
)

type Faucet struct {
	Interval                   time.Duration `yaml:"interval"`
	IntervalAddress            time.Duration `yaml:"interval_address"`
	IntervalIdentity           time.Duration `yaml:"interval_identity"`
	IntervalIdentityAndAddress time.Duration `yaml:"interval_identity_and_address"`
	IntervalIP                 time.Duration `yaml:"interval_ip"`
	Payout                     int64         `yaml:"payout"`
}

var (
	bigInt = big.NewInt(0)
	eth    = bigInt.Exp(big.NewInt(10), big.NewInt(18), nil)
)

func (f Faucet) PayoutWei() *big.Int {
	return bigInt.Mul(big.NewInt(f.Payout), eth)
}
