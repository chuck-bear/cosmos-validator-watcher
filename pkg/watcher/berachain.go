package watcher

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/kilnfi/cosmos-validator-watcher/pkg/metrics"
	"github.com/kilnfi/cosmos-validator-watcher/pkg/rpc"
	"github.com/rs/zerolog/log"
)

// Mainnet (80094) defaults from beacon-kit testing/networks/80094:
//   spec.toml  validator-set-cap=69, min-activation-balance=250_000_000_000_000
// fulu (osakaTime 1783526400 / 2026-07-08) is already live; no next fork is
// scheduled in eth-genesis.json — pass --next-upgrade-* when one is announced.
const (
	defaultValidatorSetCap      = 69
	defaultMinActivationBalance = 250_000_000_000_000 // gwei (= 250_000 BERA)
)

type BerachainWatcher struct {
	metrics *metrics.Metrics
	pool    *rpc.Pool
	opts    BerachainWatcherOptions
	client  *http.Client
}

type BerachainWatcherOptions struct {
	BeaconAPI            string
	Denom                string
	ValidatorSetCap      uint64
	MinActivationBalance uint64 // gwei
	NextUpgradeName      string
	NextUpgradeTimestamp int64 // unix seconds
}

func NewBerachainWatcher(metrics *metrics.Metrics, pool *rpc.Pool, opts BerachainWatcherOptions) *BerachainWatcher {
	if opts.Denom == "" {
		opts.Denom = "bera"
	}
	if opts.ValidatorSetCap == 0 {
		opts.ValidatorSetCap = defaultValidatorSetCap
	}
	if opts.MinActivationBalance == 0 {
		opts.MinActivationBalance = defaultMinActivationBalance
	}
	return &BerachainWatcher{
		metrics: metrics,
		pool:    pool,
		opts:    opts,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (w *BerachainWatcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		chainID := w.pool.ChainID
		w.publishStatic(chainID)

		node := w.pool.GetSyncedNode()
		if node == nil {
			log.Warn().Msg("berachain: no synced node for validators query")
		} else if err := w.fetchCometValidators(ctx, node); err != nil {
			log.Error().Err(err).Str("node", node.Redacted()).Msg("berachain: failed to fetch comet validators")
		}

		if w.opts.BeaconAPI != "" {
			if err := w.fetchBeaconStats(ctx, chainID); err != nil {
				w.metrics.BeaconAPIUp.WithLabelValues(chainID).Set(0)
				log.Error().Err(err).Str("beacon_api", w.opts.BeaconAPI).Msg("berachain: beacon api unreachable")
			}
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (w *BerachainWatcher) publishStatic(chainID string) {
	if chainID == "" {
		chainID = "unknown"
	}
	minActivationBERA := float64(w.opts.MinActivationBalance) / 1e9
	w.metrics.ValidatorSetCap.WithLabelValues(chainID).Set(float64(w.opts.ValidatorSetCap))
	w.metrics.MinActivationBalance.WithLabelValues(chainID, w.opts.Denom).Set(minActivationBERA)
	// Protocol seat price to become / stay eligible (Electra min-activation-balance).
	w.metrics.SeatPrice.WithLabelValues(chainID, w.opts.Denom).Set(minActivationBERA)
	if w.opts.NextUpgradeTimestamp > 0 && w.opts.NextUpgradeName != "" {
		w.metrics.NextUpgradeTimestamp.WithLabelValues(chainID, w.opts.NextUpgradeName).Set(float64(w.opts.NextUpgradeTimestamp))
	}
}

func (w *BerachainWatcher) fetchCometValidators(ctx context.Context, node *rpc.Node) error {
	var (
		sum   int64
		min   int64
		count int
	)

	for page := 1; page <= 50; page++ {
		perPage := 100
		p := page
		result, err := node.Client.Validators(ctx, nil, &p, &perPage)
		if err != nil {
			return fmt.Errorf("validators page %d: %w", page, err)
		}
		for _, v := range result.Validators {
			p := v.VotingPower
			sum += p
			count++
			if min == 0 || p < min {
				min = p
			}
		}
		if count >= int(result.Total) || len(result.Validators) == 0 {
			break
		}
	}

	chainID := node.ChainID()
	w.metrics.ActiveSet.WithLabelValues(chainID).Set(float64(count))
	w.metrics.TotalStake.WithLabelValues(chainID, w.opts.Denom).Set(float64(sum) / 1e9)
	if min > 0 {
		w.metrics.ActiveSetMinStake.WithLabelValues(chainID, w.opts.Denom).Set(float64(min) / 1e9)
	}

	log.Debug().
		Int("active", count).
		Uint64("cap", w.opts.ValidatorSetCap).
		Float64("total_stake_bera", float64(sum)/1e9).
		Float64("min_active_stake_bera", float64(min)/1e9).
		Float64("seat_price_bera", float64(w.opts.MinActivationBalance)/1e9).
		Int64("next_upgrade_ts", w.opts.NextUpgradeTimestamp).
		Msg("berachain stake snapshot")

	return nil
}

func (w *BerachainWatcher) fetchBeaconStats(ctx context.Context, chainID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, trimSlash(w.opts.BeaconAPI)+"/eth/v1/node/version", nil)
	if err != nil {
		return err
	}
	resp, err := w.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("beacon api status %d", resp.StatusCode)
	}
	w.metrics.BeaconAPIUp.WithLabelValues(chainID).Set(1)
	return nil
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}
