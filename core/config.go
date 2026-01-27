package core

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/devkarim/goredis/eviction"
)

var (
	ErrInvalidMaxMemory = errors.New("maxmemory must be a positive integer")
	ErrInvalidVerbose = errors.New("verbose must be a boolean")
	ErrUnknownKey       = errors.New("unknown configuration key")
)

const (
	defaultConfigFilePath = "goredis.conf"

	defaultPort      = "6379"
	defaultAofPath   = "database.aof"
	defaultMaxMemory = 1e+8
)

const (
	keyPort      = "port"
	keyAof       = "aof"
	keyMaxMemory = "maxmemory"
	keyPolicy    = "policy"
	keyVerbose   = "verbose"
)

type Config struct {
	ListenAddr *string
	AofPath    *string
	Policy     *eviction.PolicyType
	MaxMemory  *int
	Verbose    *bool
}

func ptr[T any](v T) *T { return &v }

func (cfg Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("ListenAddr", *cfg.ListenAddr),
		slog.String("AofPath", *cfg.AofPath),
		slog.Any("Policy", *cfg.Policy),
		slog.Int("MaxMemory", *cfg.MaxMemory),
		slog.Bool("Verbose", *cfg.Verbose),
	)
}

func (cfg *Config) Set(name, value string) error {
	switch name {
	case keyPort:
		cfg.ListenAddr = ptr(fmt.Sprintf(":%s", value))
	case keyAof:
		cfg.AofPath = ptr(value)
	case keyMaxMemory:
		v, err := strconv.Atoi(value)
		if err != nil || v <= 0 {
			return ErrInvalidMaxMemory
		}
		cfg.MaxMemory = ptr(v)
	case keyPolicy:
		var policy eviction.PolicyType
		if err := policy.Set(value); err != nil {
			return err
		}
		cfg.Policy = ptr(policy)
	case keyVerbose:
		v, err := strconv.ParseBool(value)
		if err != nil {
			return ErrInvalidVerbose
		}
		cfg.Verbose = ptr(v)
	default:
		return fmt.Errorf("%w: %s", ErrUnknownKey, name)
	}
	return nil
}

func (cfg *Config) Validate() error {
	if cfg.ListenAddr == nil || *cfg.ListenAddr == "" {
		return errors.New("listen address is required")
	}
	if cfg.MaxMemory == nil || *cfg.MaxMemory <= 0 {
		return ErrInvalidMaxMemory
	}
	if cfg.Policy == nil || *cfg.Policy == "" {
		return eviction.ErrInvalidPolicyType
	}
	return nil
}

func (cfg *Config) MergeWith(other Config) {
	if other.ListenAddr != nil {
		cfg.ListenAddr = other.ListenAddr
	}
	if other.AofPath != nil {
		cfg.AofPath = other.AofPath
	}
	if other.MaxMemory != nil {
		cfg.MaxMemory = other.MaxMemory
	}
	if other.Policy != nil {
		cfg.Policy = other.Policy
	}
	if other.Verbose != nil {
		cfg.Verbose = other.Verbose
	}
}

func LoadConfig() (Config, error) {
	argsConfig := LoadArgs()
	cfg, err := LoadConfigFile()

	if err != nil {
		return cfg, err
	}

	cfg.MergeWith(argsConfig)

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	slog.Info("Loaded configuration", "config", cfg)

	return cfg, nil
}

func LoadDefaultConfig() Config {
	return Config{
		ListenAddr: ptr(fmt.Sprintf(":%s", defaultPort)),
		AofPath:    ptr(defaultAofPath),
		MaxMemory:  ptr(int(defaultMaxMemory)),
		Policy:     ptr(eviction.PolicyLRU),
		Verbose:    ptr(false),
	}
}

func LoadArgs() Config {
	var policy eviction.PolicyType
	flag.String(keyPort, defaultPort, "Port to listen on")
	flag.String(keyAof, defaultAofPath, "Path of the AOF file")
	flag.Int(keyMaxMemory, defaultMaxMemory, "Max memory in bytes")
	flag.Var(&policy, keyPolicy, "Eviction policy: lru, fifo")
	flag.Bool(keyVerbose, false, "Verbose mode for debugging")

	flag.Parse()

	cfg := Config{}

	flag.Visit(func(f *flag.Flag) {
		if err := cfg.Set(f.Name, f.Value.String()); err != nil {
			slog.Warn("invalid flag value", "flag", f.Name, "error", err)
		}
	})

	return cfg
}

func LoadConfigFile() (Config, error) {
	cfg := LoadDefaultConfig()
	file, err := os.Open(defaultConfigFilePath)

	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			slog.Info("No configuration file detected")
			return cfg, nil
		}

		return Config{}, err
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		parts := strings.SplitN(text, " ", 2)
		if len(parts) == 2 {
			if err := cfg.Set(parts[0], strings.TrimSpace(parts[1])); err != nil {
				slog.Warn("invalid config value", "key", parts[0], "error", err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}
