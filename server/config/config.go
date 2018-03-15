package config

import (
	"../ses"
	"path/filepath"
	"io/ioutil"
	"gopkg.in/yaml.v2"
)

// config file structure
type config struct {
	Debug bool `yaml:"Debug"`

	AwsKey    string `yaml:"AwsKey"`
	AwsSecret string `yaml:"AwsSecret"`
	AwsRegion string `yaml:"AwsRegion"`

	NoReplyEmail string `yaml:"NoReplyEmail"`
	ReplyEmail   string `yaml:"ReplyEmail"`

	DatabaseDriver string `yaml:"DatabaseDriver"`
	DatabaseDSN    string `yaml:"DatabaseDSN"`

	MaxFileUploadSizeMb int64 `yaml:"MaxFileUploadSizeMb"`

	Port string `yaml:"Port"`
}

var Config = &config{}

func init() {
	loadConfig()

	// Amazon SES setup
	ses.SetConfiguration(Config.AwsKey, Config.AwsSecret, Config.AwsRegion)
}

func loadConfig() {
	// which will try to find the 'filename' from current working dir too.
	yamlAbsPath, err := filepath.Abs("config.yml")
	if err != nil {
		println("Can't find example.config.yml " + err.Error())
	}

	// read the raw contents of the file
	data, err := ioutil.ReadFile(yamlAbsPath)
	if err != nil {
		println("Can't read example.config.yml " + err.Error())
	}

	err = yaml.Unmarshal(data, &Config)
	if err != nil {
		panic(err)
	}
}
