package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
)

// LoadConfig loads a WarpConfig from given filepath
func LoadConfig(configPath string) (*Config, error) {
	raw, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err = json.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}

	//	cfg.setDefaults()

	return &cfg, nil
}

// GenerateConfig writes a empty TemporalConfig template to given filepath
func GenerateConfig(configPath string) error {
	template := &Config{}
	//template.setDefaults()
	b, err := json.Marshal(template)
	if err != nil {
		return err
	}

	var pretty bytes.Buffer
	if err = json.Indent(&pretty, b, "", "\t"); err != nil {
		return err
	}
	return ioutil.WriteFile(configPath, pretty.Bytes(), os.ModePerm)
}
