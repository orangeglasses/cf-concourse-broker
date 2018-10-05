package main

import (
	"errors"
	"fmt"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc"
	"github.com/concourse/go-concourse/concourse"
)

const adminTeam = "main"

// IccClient defines the capabilities that any concourse client should be able to do.
type IccClient interface {
	CreateTeam(details cfDetails) error
	DeleteTeam(details cfDetails) error
}

// NewClient returns a client that can be used to interface with a deployed Concourse CI instance.
func concourseNewClient(env brokerConfig, logger lager.Logger) IccClient {
	httpClient := newBasicAuthClient(env.AdminUsername, env.AdminPassword)

	return &concourseClient{
		client: concourse.NewClient(env.ConcourseURL, httpClient, false),
		env:    env,
		logger: logger.Session("concourse-client")}
}

type concourseClient struct {
	client concourse.Client
	env    brokerConfig
	logger lager.Logger
}

func (c *concourseClient) getAuthClient(concourseURL string) (concourse.Client, error) {
	return c.client, nil
}

func (c *concourseClient) getTeamName(details cfDetails) string {
	return details.OrgName
}

func (c *concourseClient) CreateTeam(details cfDetails) error {
	teamName := c.getTeamName(details)

	orgSpaceAuth := fmt.Sprintf("%v:%v", details.OrgName, details.SpaceName)
	team := atc.Team{}
	team.Auth["groups"] = []string{orgSpaceAuth}

	client, err := c.getAuthClient(c.env.ConcourseURL)
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
	client, err := c.getAuthClient(c.env.ConcourseURL)
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
