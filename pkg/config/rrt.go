package config

import "time"

type RRTConfig struct {
	Verbose bool
	Pretty  bool
	Format  string

	Kubeconfig string
	Namespace  string

	Timeout time.Duration
	Grace   time.Duration
}
