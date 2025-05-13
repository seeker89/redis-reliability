package config

import "time"

type RRConfig struct {
	Verbose bool
	Pretty  bool
	Format  string

	Kubeconfig string
	Namespace  string

	Timeout time.Duration
	Grace   time.Duration

	SentinelURL    string
	SentinelMaster string
}
