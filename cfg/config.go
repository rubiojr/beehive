package cfg

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"strings"

	"github.com/muesli/beehive/bees"
	"github.com/rubiojr/fcrypto"
	log "github.com/sirupsen/logrus"
)

// Config contains an entire configuration set for Beehive
type Config struct {
	Bees    []bees.BeeConfig
	Actions []bees.Action
	Chains  []bees.Chain
}

// Loads chains from config
func LoadConfig(file string) (Config, error) {
	var config Config

	var j []byte
	var err error
	var buf *bytes.Buffer
	secret := os.Getenv("BEEHIVE_CONFIG_SECRET")
	if secret == "" && isConfigEncrypted(file) {
		log.Fatalf("Configuration file '%s' is encrypted, but the BEEHIVE_CONFIG_SECRET environment variable was not provided", file)
	}
	if secret != "" && !isConfigEncrypted(file) {
		log.Info("Secret provided but the config is not encrypted. Encrypting configuration...")
	}
	if secret != "" {
		log.Println("Loading encrypted config file")
		buf, err = fcrypto.LoadFile(file, secret)
		if err == nil {
			j = buf.Bytes()
		}
	} else {
		j, err = ioutil.ReadFile(file)
	}

	if err != nil {
		return config, err
	}

	config = Config{}
	err = json.Unmarshal(j, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

// Saves chains to config
func SaveConfig(file string, c Config) error {
	secret := os.Getenv("BEEHIVE_CONFIG_SECRET")
	j, err := json.MarshalIndent(c, "", "  ")
	if err == nil {
		if secret != "" {
			log.Println("Saving encrypted config file")
			fcrypto.SaveFile(bytes.NewBuffer(j), file, secret)
		} else {
			err = ioutil.WriteFile(file, j, 0644)
		}
	}

	return err
}

func SaveCurrentConfig(file string) error {
	config := Config{}
	config.Bees = bees.BeeConfigs()
	config.Chains = bees.GetChains()
	config.Actions = bees.GetActions()
	return SaveConfig(file, config)
}

func Exist(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}

	return true
}

func isConfigEncrypted(file string) bool {
	f, err := os.Open(file)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Scan()
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return strings.HasPrefix(scanner.Text(), "# Encrypted fcrypto")
}
