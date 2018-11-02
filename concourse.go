package main

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"time"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/go-concourse/concourse"
	"golang.org/x/oauth2"
)

const adminTeam = "main"

// IccClient defines the capabilities that any concourse client should be able to do.
type IccClient interface {
	CreateTeam(details cfDetails) error
	DeleteTeam(details cfDetails) error
}

type concourseClient struct {
	client concourse.Client
	env    brokerConfig
	logger lager.Logger
}

// NewClient returns a client that can be used to interface with a deployed Concourse CI instance.
func concourseNewClient(env brokerConfig, logger lager.Logger) IccClient {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).Dial,
		},
	}

	return &concourseClient{
		client: concourse.NewClient(env.ConcourseURL, httpClient, false),
		env:    env,
		logger: logger.Session("concourse-client")}
}

func (c *concourseClient) getAuthClient() (concourse.Client, error) {
	oauth2Config := oauth2.Config{
		ClientID:     "fly",
		ClientSecret: "Zmx5",
		Endpoint:     oauth2.Endpoint{TokenURL: c.client.URL() + "/sky/token"},
		Scopes:       []string{"openid", "profile", "email", "federated:id", "groups"},
	}

	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, c.client.HTTPClient())

	token, err := oauth2Config.PasswordCredentialsToken(ctx, c.env.AdminUsername, c.env.AdminPassword)
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

	return concourse.NewClient(c.env.ConcourseURL, httpClient, false), nil
}

func (c *concourseClient) getTeamName(details cfDetails) string {
	return details.OrgName
}

func (c *concourseClient) CreateTeam(details cfDetails) error {
	teamName := c.getTeamName(details)
	team := atc.Team{
		Name: teamName,
		Auth: atc.TeamAuth{
			"owner": map[string][]string{
				"groups": []string{details.OrgName},
			},
		},
	}

	client, err := c.getAuthClient()
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

func (c *concourseClient) DeleteTeam(details cfDetails) error {
	teamName := c.getTeamName(details)
	client, err := c.getAuthClient()
	if err != nil {
		c.logger.Error("delete-team.auth-client-error", err)
		return err
	}
	err = client.Team(details.OrgName).DestroyTeam(teamName)
	if err != nil {
		c.logger.Error("delete-team.unknown-delete-error", err,
			lager.Data{
				"team-name": teamName,
			})
		return err
	}
	return nil
}
