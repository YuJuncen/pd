// Copyright 2022 TiKV Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tso

import (
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pingcap/errors"
	"github.com/pingcap/log"
	"github.com/spf13/pflag"
	"github.com/tikv/pd/pkg/encryption"
	"github.com/tikv/pd/pkg/utils/grpcutil"
	"github.com/tikv/pd/pkg/utils/metricutil"
	"github.com/tikv/pd/pkg/utils/typeutil"
	"go.uber.org/zap"
)

const (
	// defaultTSOUpdatePhysicalInterval is the default value of the config `TSOUpdatePhysicalInterval`.
	defaultTSOUpdatePhysicalInterval = 50 * time.Millisecond
)

// Config is the configuration for the TSO.
type Config struct {
	BackendEndpoints string `toml:"backend-endpoints" json:"backend-endpoints"`
	ListenAddr       string `toml:"listen-addr" json:"listen-addr"`

	// EnableLocalTSO is used to enable the Local TSO Allocator feature,
	// which allows the PD server to generate Local TSO for certain DC-level transactions.
	// To make this feature meaningful, user has to set the "zone" label for the PD server
	// to indicate which DC this PD belongs to.
	EnableLocalTSO bool `toml:"enable-local-tso" json:"enable-local-tso"`

	// TSOSaveInterval is the interval to save timestamp.
	TSOSaveInterval typeutil.Duration `toml:"tso-save-interval" json:"tso-save-interval"`

	// The interval to update physical part of timestamp. Usually, this config should not be set.
	// At most 1<<18 (262144) TSOs can be generated in the interval. The smaller the value, the
	// more TSOs provided, and at the same time consuming more CPU time.
	// This config is only valid in 1ms to 10s. If it's configured too long or too short, it will
	// be automatically clamped to the range.
	TSOUpdatePhysicalInterval typeutil.Duration `toml:"tso-update-physical-interval" json:"tso-update-physical-interval"`

	// MaxResetTSGap is the max gap to reset the TSO.
	MaxResetTSGap typeutil.Duration `toml:"max-gap-reset-ts" json:"max-gap-reset-ts"`

	Metric metricutil.MetricConfig `toml:"metric" json:"metric"`

	// Log related config.
	Log log.Config `toml:"log" json:"log"`

	Logger   *zap.Logger
	LogProps *log.ZapProperties

	Security SecurityConfig `toml:"security" json:"security"`
}

// NewConfig creates a new config.
func NewConfig() *Config {
	return &Config{}
}

// Parse parses flag definitions from the argument list.
func (c *Config) Parse(flagSet *pflag.FlagSet) error {
	// Load config file if specified.
	if configFile, _ := flagSet.GetString("config"); configFile != "" {
		_, err := c.configFromFile(configFile)
		if err != nil {
			return err
		}
	}

	// ignore the error check here
	adjustCommandlineString(flagSet, &c.Log.Level, "log-level")
	adjustCommandlineString(flagSet, &c.Log.File.Filename, "log-file")
	adjustCommandlineString(flagSet, &c.Metric.PushAddress, "metrics-addr")
	adjustCommandlineString(flagSet, &c.Security.CAPath, "cacert")
	adjustCommandlineString(flagSet, &c.Security.CertPath, "cert")
	adjustCommandlineString(flagSet, &c.Security.KeyPath, "key")
	adjustCommandlineString(flagSet, &c.BackendEndpoints, "backend-endpoints")
	adjustCommandlineString(flagSet, &c.ListenAddr, "listen-addr")

	// TODO: Implement the main function body
	return nil
}

// configFromFile loads config from file.
func (c *Config) configFromFile(path string) (*toml.MetaData, error) {
	meta, err := toml.DecodeFile(path, c)
	return &meta, errors.WithStack(err)
}

// SecurityConfig indicates the security configuration for pd server
type SecurityConfig struct {
	grpcutil.TLSConfig
	// RedactInfoLog indicates that whether enabling redact log
	RedactInfoLog bool              `toml:"redact-info-log" json:"redact-info-log"`
	Encryption    encryption.Config `toml:"encryption" json:"encryption"`
}

func adjustCommandlineString(flagSet *pflag.FlagSet, v *string, name string) {
	if value, _ := flagSet.GetString(name); value != "" {
		*v = value
	}
}
