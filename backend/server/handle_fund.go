package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/flashbots/eth-faucet/logutils"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var (
	ErrAuthorisationHeaderMalformed = errors.New("authorisation header is malformed")
	ErrAuthorisationHeaderMissing   = errors.New("authorisation header is missing")
	ErrJWTFailedToParse             = errors.New("failed to parse jwt token")
	ErrJWTInvalidSchema             = errors.New("jwt token has unrecognised schema")
	ErrRatelimiterTooFewProxies     = errors.New("too few proxies")
	ErrRequestFailedToParse         = errors.New("failed to parse request body")
	ErrRequestFailedToRead          = errors.New("failed to read request body")
)

type requestFund struct {
	Address string `json:"address"`
}

type responseFund struct {
	Message string `json:"message"`
}

type jwtFund struct {
	jwt.RegisteredClaims

	Provider string `json:"provider"`
	Username string `json:"username"`
}

func (s *Server) handleFund(w http.ResponseWriter, r *http.Request) {
	l := logutils.LoggerFromRequest(r)

	if r.Method != http.MethodPost {
		s.httpError(w, http.StatusNotImplemented)
		return
	}

	claims, err := s.authoriseRequestFund(r)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			http.Error(
				w,
				"Error: your api token is expired, please try refreshing the page",
				http.StatusForbidden,
			)
			return
		}

		l.Warn("Failed to authorise fund request", zap.Error(err))
		s.httpError(w, http.StatusForbidden)
		return
	}

	request, err := s.parseRequestFund(r)
	if err != nil {
		l.Warn("Failed to parse fund request", zap.Error(err))
		s.httpError(w, http.StatusBadRequest)
		return
	}

	wait, err := s.ratelimitRequestFund(r, claims, request)
	if err != nil {
		l.Warn("Failed to rate-limit fund request", zap.Error(err))
		s.httpError(w, http.StatusTooManyRequests)
		return
	}
	if wait > time.Duration(0) {
		err = s.renderJSON(w, http.StatusTooManyRequests, &responseFund{
			Message: "Too many requests, come back in " + wait.Round(time.Second).String(),
		})
		if err != nil {
			l := logutils.LoggerFromRequest(r)
			l.Error("Failed to send fund response", zap.Error(err))
		}
		return
	}

	txHash, err := s.txbuilder.SendFunds(r.Context(), request.Address, s.cfg.Faucet.PayoutWei())
	if err != nil {
		l.Error("Failed to send funds",
			zap.Error(err),
			zap.Int64("amount", s.cfg.Faucet.Payout),
			zap.String("address_from", s.txbuilder.Address()),
			zap.String("address_to", request.Address),
			zap.String("identity_provider", claims.Provider),
			zap.String("identity_username", claims.Username),
			zap.String("tx_hash", txHash.Hex()),
		)
		err = s.renderJSON(w, http.StatusOK, &responseFund{
			Message: "Error: " + err.Error(),
		})
		if err != nil {
			l := logutils.LoggerFromRequest(r)
			l.Error("Failed to send fund response", zap.Error(err))
		}
		return
	}

	l.Info("Sent funds",
		zap.Int64("amount", s.cfg.Faucet.Payout),
		zap.String("address_from", s.txbuilder.Address()),
		zap.String("address_to", request.Address),
		zap.String("identity_provider", claims.Provider),
		zap.String("identity_username", claims.Username),
		zap.String("tx_hash", txHash.Hex()),
	)

	err = s.renderJSON(w, http.StatusOK, &responseFund{
		Message: "TxHash: " + txHash.Hex(),
	})
	if err != nil {
		l := logutils.LoggerFromRequest(r)
		l.Error("Failed to send fund response", zap.Error(err))
	}
}

func (s *Server) authoriseRequestFund(r *http.Request) (
	*jwtFund, error,
) {
	authorizationHeader := r.Header.Get("authorization")
	if authorizationHeader == "" {
		return nil, ErrAuthorisationHeaderMissing
	}
	if !strings.HasPrefix(authorizationHeader, "Bearer ") {
		return nil, ErrAuthorisationHeaderMalformed
	}
	tokenString := strings.TrimPrefix(authorizationHeader, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &jwtFund{}, func(_ *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.Server.AuthSecret), nil
	}, jwt.WithLeeway(5*time.Second))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrJWTFailedToParse, err)
	}
	claims, ok := token.Claims.(*jwtFund)
	if !ok {
		return nil, ErrJWTInvalidSchema
	}
	return claims, nil
}

func (s *Server) parseRequestFund(r *http.Request) (
	*requestFund, error,
) {
	body, err := io.ReadAll(io.LimitReader(r.Body, int64(s.cfg.Server.MaxRequestBodySize)))
	defer r.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestFailedToRead, err)
	}

	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()

	request := &requestFund{}
	if err := d.Decode(request); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrRequestFailedToParse, err)
	}

	r.Body = io.NopCloser(bytes.NewReader(body))
	return request, nil
}

func (s *Server) ratelimitRequestFund(
	r *http.Request,
	claims *jwtFund,
	request *requestFund,
) (
	time.Duration, error,
) {
	forwardedFor := strings.Split(r.Header.Get("x-forwarded-for"), ",")
	if len(forwardedFor) < s.cfg.Server.ProxyCount {
		return time.Duration(0), fmt.Errorf("%w: %d", ErrRatelimiterTooFewProxies, len(forwardedFor))
	}
	entryIP := strings.TrimSpace(forwardedFor[len(forwardedFor)-1-s.cfg.Server.ProxyCount])

	ratelimitKeys := map[string]time.Duration{
		fmt.Sprintf("address:%s", request.Address):                                      max(s.cfg.Faucet.Interval, s.cfg.Faucet.IntervalAddress),
		fmt.Sprintf("full:%s:%s:%s", claims.Provider, claims.Username, request.Address): max(s.cfg.Faucet.Interval, s.cfg.Faucet.IntervalIdentityAndAddress),
		fmt.Sprintf("identity:%s:%s", claims.Provider, claims.Username):                 max(s.cfg.Faucet.Interval, s.cfg.Faucet.IntervalIdentity),
		fmt.Sprintf("ip:%s", entryIP):                                                   max(s.cfg.Faucet.Interval, s.cfg.Faucet.IntervalIP),
	}

	nextAllowed := time.Now()
	for key, interval := range ratelimitKeys {
		timestamp, err := s.ratelimiter.IsRegistered(r.Context(), key)
		if err != nil {
			return time.Duration(0), err
		}
		timestamp = timestamp.Add(interval)
		if timestamp.After(nextAllowed) {
			nextAllowed = timestamp
		}
	}
	interval := time.Until(nextAllowed)

	if interval > time.Duration(0) {
		return interval, nil
	}

	for key, expiry := range ratelimitKeys {
		if strings.HasPrefix(key, "full:") { // debug
			expiry = 24 * time.Hour
		}
		if err := s.ratelimiter.Register(r.Context(), key, expiry); err != nil {
			return time.Duration(0), err
		}
	}

	return time.Duration(0), nil
}
