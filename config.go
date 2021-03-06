package main

import "github.com/kelseyhightower/envconfig"

type brokerConfig struct {
	BrokerUsername string `envconfig:"broker_username" required:"true"`
	BrokerPassword string `envconfig:"broker_password" required:"true"`
	AdminUsername  string `envconfig:"admin_username" required:"true"`
	AdminPassword  string `envconfig:"admin_password" required:"true"`
	ConcourseURL   string `envconfig:"concourse_url" required:"true"`
	CFURL          string `envconfig:"cf_url" required:"true"`
	TokenURL       string `envconfig:"token_url" required:"true"`
	AuthURL        string `envconfig:"auth_url" required:"true"`
	ClientID       string `envconfig:"client_id" required:"true"`
	ClientSecret   string `envconfig:"client_secret" required:"true"`
	LogLevel       string `envconfig:"log_level" default:"INFO"`
	Port           string `envconfig:"port" default:"3000"`
}

func brokerConfigLoad() (brokerConfig, error) {
	var config brokerConfig
	err := envconfig.Process("", &config)
	if err != nil {
		return brokerConfig{}, err
	}

	return config, nil
}
