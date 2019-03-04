package main

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc"
	"github.com/concourse/go-concourse/concourse"
	"golang.org/x/oauth2"
)

type concourseTarget struct {
	client   concourse.Client
	username string
	password string
	logger   lager.Logger
}

func concourseNewClient(env brokerConfig, logger lager.Logger) *concourseTarget {
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			Dial: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).Dial,
		},
	}

	return &concourseTarget{
		client:   concourse.NewClient(env.ConcourseURL, httpClient, false),
		username: env.AdminUsername,
		password: env.AdminPassword,
		logger:   logger.Session("concourse-target")}
}

func (c *concourseTarget) Client() (concourse.Client, error) {
	oauth2Config := oauth2.Config{
		ClientID:     "fly",
		ClientSecret: "Zmx5",
		Endpoint:     oauth2.Endpoint{TokenURL: c.client.URL() + "/sky/token"},
		Scopes:       []string{"openid", "profile", "email", "federated:id", "groups"},
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, c.client.HTTPClient())

	token, err := oauth2Config.PasswordCredentialsToken(ctx, c.username, c.password)
	if err != nil {
		return nil, err
	}

	var transport http.RoundTripper

	transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		Dial: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).Dial,
		Proxy: http.ProxyFromEnvironment,
	}

	httpClient := &http.Client{
		Transport: &oauth2.Transport{
			Source: oauth2.StaticTokenSource(token),
			Base:   transport,
		},
	}

	return concourse.NewClient(c.client.URL(), httpClient, false), nil
}

func (c *concourseTarget) CreateTeam(orgName string) error {
	teamName := orgName
	groups := []string{}
	users := []string{}

	groups = append(groups, "cf:"+strings.ToLower(orgName))

	team := atc.Team{
		Auth: map[string][]string{
			"users":  users,
			"groups": groups,
		},
	}

	client, err := c.Client()
	if err != nil {
		c.logger.Error("create-team.auth-client-error", err)
		return err
	}

	_, created, updated, err := client.Team(teamName).CreateOrUpdate(team)
	if err != nil {
		c.logger.Error("create-team.unknown-create-error", err,
			lager.Data{
				"team-name": teamName,
			})
		return err
	}
	if !created || updated {
		err := errors.New("Unable to provision instance")
		c.logger.Error("create-team.unknown-create-error", err,
			lager.Data{
				"team-name": teamName,
			})
		return err
	}
	return nil
}

func (c *concourseTarget) DeleteTeam(teamName string) error {
	client, err := c.Client()
	if err != nil {
		c.logger.Error("delete-team.auth-client-error", err)
		return err
	}
	err = client.Team(teamName).DestroyTeam(teamName)
	if err != nil {
		c.logger.Error("delete-team.unknown-delete-error", err,
			lager.Data{
				"team-name": teamName,
			})
		return err
	}
	return nil
}
