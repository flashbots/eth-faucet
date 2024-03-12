package config_test

import (
	"testing"

	"github.com/flashbots/eth-faucet/config"
	"github.com/stretchr/testify/assert"
)

func TestPayoutWei(t *testing.T) {
	f := config.Faucet{
		Payout: 10,
	}
	val1, _ := f.PayoutWei().Float64()
	assert.Equal(t, 1e+19, val1)
	val2, _ := f.PayoutWei().Float64()
	assert.Equal(t, 1e+19, val2)
}
