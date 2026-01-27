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
)

var ErrInvalidPolicy = errors.New("invalid policy, must be one of: lru, fifo")

var (
	defaultConfigFilePath = "goredis.conf"

	defaultPort                    = "6379"
	defaultAofPath                 = "database.aof"
	defaultPolicy    AllowedPolicy = "lru"
	defaultMaxMemory int           = 1e+8
)

type AllowedPolicy string

func (p *AllowedPolicy) String() string {
	return string(*p)
}

func (p *AllowedPolicy) Set(value string) error {
	switch value {
	case "fifo", "lru":
		*p = AllowedPolicy(value)
		return nil
	default:
		return ErrInvalidPolicy
	}
}

type Config struct {
	ListenAddr string
	AofPath    string
	Policy     string
	MaxMemory  int
}

func (cfg *Config) Set(name, value string) {
	switch name {
	case "port":
		cfg.ListenAddr = fmt.Sprintf(":%s", value)
	case "aof":
		cfg.AofPath = value
	case "maxmemory":
		value, err := strconv.Atoi(value)
		if err != nil {
			slog.Warn("maxmemory is not an integer")
			return
		}
		cfg.MaxMemory = value
	case "policy":
		var policy *AllowedPolicy
		err := policy.Set(value)
		if err != nil {
			slog.Warn("invalid value for policy")
			return
		}
		cfg.Policy = value
	}
}

func (cfg *Config) MergeWith(anotherConfig Config) Config {
	if anotherConfig.ListenAddr != "" {
		cfg.ListenAddr = anotherConfig.ListenAddr
	}
	if anotherConfig.AofPath != "" {
		cfg.AofPath = anotherConfig.AofPath
	}
	if anotherConfig.MaxMemory != 0 {
		cfg.MaxMemory = anotherConfig.MaxMemory
	}
	if anotherConfig.Policy != "" {
		cfg.Policy = anotherConfig.Policy
	}
	return *cfg
}

func LoadConfig() (Config, error) {
	argsConfig := LoadArgs()
	cfg, err := LoadConfigFile()

	if err != nil {
		return cfg, err
	}

	cfg = cfg.MergeWith(argsConfig)

	slog.Info("Loaded configuration", "config", cfg)

	return cfg, nil
}

func LoadDefaultConfig() Config {
	return Config{
		ListenAddr: fmt.Sprintf(":%s", defaultPort),
		AofPath:    defaultAofPath,
		MaxMemory:  int(defaultMaxMemory),
		Policy:     string(defaultPolicy),
	}
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func LoadArgs() Config {
	var policy AllowedPolicy = defaultPolicy
	port := flag.String("port", defaultPort, "Port to listen on")
	aofPath := flag.String("aof", defaultAofPath, "Path of the AOF file")
	maxMemory := flag.Int("maxmemory", defaultMaxMemory, "Max memory in bytes")
	flag.Var(&policy, "policy", "Eviction policy: lru, fifo")

	flag.Parse()

	cfg := Config{}

	if isFlagPassed("port") {
		cfg.Set("port", *port)
	}

	if isFlagPassed("aof") {
		cfg.Set("aof", *aofPath)
	}

	if isFlagPassed("policy") {
		cfg.Set("policy", string(policy))
	}

	if isFlagPassed("maxmemory") {
		cfg.Set("maxmemory", strconv.Itoa(*maxMemory))
	}

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
		line := strings.Split(scanner.Text(), " ")
		if len(line) >= 2 {
			key, value := line[0], line[1]
			cfg.Set(key, value)
		}
	}
	return cfg, nil
}
