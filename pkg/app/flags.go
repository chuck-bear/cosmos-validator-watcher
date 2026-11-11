package app

import (
	"sort"
	"time"

	"github.com/urfave/cli/v2"
)

var Flags = []cli.Flag{
	&cli.StringFlag{
		Name:  "chain-id",
		Usage: "to ensure all nodes matches the specific network (dismiss to auto-detected)",
	},
	&cli.BoolFlag{
		Name:  "debug",
		Usage: "shortcut for --log-level=debug",
	},
	&cli.StringFlag{
		Name:  "http-addr",
		Usage: "http server address",
		Value: ":8080",
	},
	&cli.StringFlag{
		Name:  "log-level",
		Usage: "log level (debug, info, warn, error)",
		Value: "info",
	},
	&cli.StringFlag{
		Name:  "namespace",
		Usage: "namespace for Prometheus metrics",
		Value: "cosmos_validator_watcher",
	},
	&cli.BoolFlag{
		Name:  "no-color",
		Usage: "disable colored output",
	},
	&cli.StringSliceFlag{
		Name:  "node",
		Usage: "rpc node endpoint to connect to (specify multiple for high availability)",
		Value: cli.NewStringSlice("http://localhost:26657"),
	},
	&cli.BoolFlag{
		Name:  "no-gov",
		Usage: "disable calls to gov module (useful for consumer chains)",
	},
	&cli.BoolFlag{
		Name:  "no-staking",
		Usage: "disable calls to staking module (useful for consumer chains)",
	},
	&cli.BoolFlag{
		Name:  "no-slashing",
		Usage: "disable calls to slashing module",
	},
	&cli.BoolFlag{
		Name:  "no-commission",
		Usage: "disable calls to get validator commission (useful for chains without distribution module)",
	},
	&cli.BoolFlag{
		Name:  "no-upgrade",
		Usage: "disable calls to upgrade module (for chains created without the upgrade module)",
	},
	&cli.StringFlag{
		Name:  "denom",
		Usage: "denom used in metrics label (eg. atom or uatom)",
	},
	&cli.UintFlag{
		Name:  "denom-exponent",
		Usage: "denom exponent (eg. 6 for atom, 1 for uatom)",
	},
	&cli.DurationFlag{
		Name:  "start-timeout",
		Usage: "timeout to wait on startup for one node to be ready",
		Value: 10 * time.Second,
	},
	&cli.DurationFlag{
		Name:  "stop-timeout",
		Usage: "timeout to wait on stop",
		Value: 10 * time.Second,
	},
	&cli.StringSliceFlag{
		Name:  "validator",
		Usage: "validator address(es) to track (use :my-label to add a custom label in metrics & output)",
	},
	&cli.StringFlag{
		Name:  "webhook-url",
		Usage: "endpoint where to send upgrade webhooks (experimental)",
	},
	&cli.StringSliceFlag{
		Name:  "webhook-custom-block",
		Usage: "trigger a custom webhook at a given block number (experimental)",
	},
	&cli.StringFlag{
		Name:  "x-gov",
		Usage: "version of the gov module to use (v1|v1beta1)",
		Value: "v1",
	},
	&cli.BoolFlag{
		Name:  "babylon",
		Usage: "enable babylon watcher (checkpoint votes & finality providers)",
	},
	&cli.StringSliceFlag{
		Name:  "finality-provider",
		Usage: "list of finality providers to watch (requires --babylon)",
	},
	&cli.BoolFlag{
		Name:  "berachain",
		Usage: "enable berachain watcher (total stake, seat price, validator cap, next upgrade)",
	},
	&cli.StringFlag{
		Name:  "beacon-api",
		Usage: "beacon HTTP API base URL (requires --berachain), e.g. http://localhost:3500",
		Value: "http://localhost:3500",
	},
	&cli.Uint64Flag{
		Name:  "validator-set-cap",
		Usage: "berachain validator-set-cap from chain spec (default: mainnet 69)",
		Value: 69,
	},
	&cli.Uint64Flag{
		Name:  "min-activation-balance",
		Usage: "berachain min-activation-balance in gwei (default: 250000 BERA = 250000000000000)",
		Value: 250_000_000_000_000,
	},
	&cli.StringFlag{
		Name:  "next-upgrade-name",
		Usage: "label for next berachain upgrade metric (default: fulu)",
		Value: "fulu",
	},
	&cli.Int64Flag{
		Name:  "next-upgrade-timestamp",
		Usage: "unix timestamp of next berachain upgrade (default: fulu/osakaTime 1783526400)",
		Value: 1_783_526_400,
	},
}

func init() {
	sort.SliceStable(Flags, func(i, j int) bool {
		return Flags[i].Names()[0] < Flags[j].Names()[0]
	})
}
