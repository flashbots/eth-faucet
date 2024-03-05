package server

import (
	"net/http"
	"strconv"

	"github.com/flashbots/eth-faucet/logutils"
	"go.uber.org/zap"
)

type responseInfo struct {
	Address string `json:"address"`
	Chain   string `json:"network"`
	Payout  string `json:"payout"`
	Symbol  string `json:"symbol"`
}

func (s *Server) handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.httpError(w, http.StatusNotImplemented)
		return
	}

	err := s.renderJSON(w, http.StatusOK, responseInfo{
		Address: s.txbuilder.Address(),
		Chain:   s.cfg.Chain.Name,
		Payout:  strconv.FormatInt(s.cfg.Faucet.Payout, 10),
		Symbol:  s.cfg.Chain.TokenSymbol,
	})
	if err != nil {
		l := logutils.LoggerFromRequest(r)
		l.Error("Failed to send info response", zap.Error(err))
	}
}
