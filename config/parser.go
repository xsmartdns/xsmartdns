package config

import (
	"encoding/json"
	"fmt"
)

func Parse(b []byte) (*Config, error) {
	cfg := &Config{}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, err
	}
	cfg.FillDefault()
	if err := cfg.Verify(); err != nil {
		return nil, fmt.Errorf("verify config err:%v", err)
	}
	return cfg, nil
}
