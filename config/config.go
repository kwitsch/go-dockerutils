package config

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/creasty/defaults"
	"github.com/knadh/koanf"
	"github.com/knadh/koanf/providers/env"
)

const secretsDir string = "/run/secrets"

// initialize struct with givven prefix from environment and docker secrets
func Load(prefix string, config interface{}) error {
	err := defaults.Set(config)
	if err == nil {
		if !strings.HasSuffix(prefix, "_") {
			prefix += "_"
		}
		loadSecrets(prefix)
		var k = koanf.New(".")
		k.Load(env.Provider(prefix, ".", func(s string) string {
			return strings.Replace(strings.ToLower(
				strings.TrimPrefix(s, prefix)), "_", ".", -1)
		}), nil)
		err = k.UnmarshalWithConf("", config, koanf.UnmarshalConf{Tag: "koanf"})
	}
	return err
}

// loads allsecrets with prefix to environment
func loadSecrets(prefix string) {
	if _, ferr := os.Stat(secretsDir); !os.IsNotExist(ferr) {
		files, ferr := ioutil.ReadDir(secretsDir)
		if ferr == nil {
			for _, file := range files {
				if strings.HasPrefix(file.Name(), prefix) && !file.IsDir() && file.Size() > 0 {
					bval, eerr := ioutil.ReadFile(secretsDir + "/" + file.Name())
					if eerr == nil {
						sval := strings.TrimSpace(string(bval))
						keyname := strings.ToUpper(file.Name())
						os.Setenv(keyname, sval)
					}
				}
			}
		}
	}
}
