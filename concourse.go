package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"code.cloudfoundry.org/lager"
	"github.com/concourse/atc"
	"github.com/concourse/atc/auth/uaa"
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
		client: concourse.NewClient(env.ConcourseURL, httpClient),
		env:    env,
		logger: logger.Session("concourse-client")}
}

type concourseClient struct {
	client concourse.Client
	env    brokerConfig
	logger lager.Logger
}

func (c *concourseClient) getAuthClient(concourseURL string) (concourse.Client, error) {
	team := c.client.Team(adminTeam)
	token, err := team.AuthToken()
	if err != nil {
		return nil, err
	}
	httpClient := newOAuthClient(token.Type, token.Value)
	return concourse.NewClient(concourseURL, httpClient), nil
}

func (c *concourseClient) getTeamName(details cfDetails) string {
	return details.OrgName
}

func (c *concourseClient) CreateTeam(details cfDetails) error {
	teamName := c.getTeamName(details)
	team := atc.Team{}
	teamAuth := make(map[string]*json.RawMessage)

	uaaAuthConfig := uaa.UAAAuthConfig{
		ClientID:     c.env.ClientID,
		ClientSecret: c.env.ClientSecret,
		AuthURL:      c.env.AuthURL,
		TokenURL:     c.env.TokenURL,
		CFSpaces:     []string{details.SpaceGUID},
		CFCACert:     "",
		CFURL:        c.env.CFURL,
	}

	data, err := json.Marshal(uaaAuthConfig)
	if err != nil {
		fmt.Println("Invalid UAA config")
		panic(err)
	}

	teamAuth["uaa"] = (*json.RawMessage)(&data)
	/*fmt.Printf("{ClientID:%v,ClientSecret:%v,AuthURL:%v,TokenURL:%v,CFSpaces:%v,CFCACert:\"\",CFURL:%v}", c.env.ClientID, c.env.ClientSecret, c.env.AuthURL, c.env.TokenURL, []string{details.SpaceGUID}, c.env.CFURL)*/

	team.Auth = teamAuth

	client, err := c.getAuthClient(c.env.ConcourseURL)
	if err != nil {
		c.logger.Error("create-team.auth-client-error", err)
		return err
	}
	authMethods, err := client.Team(teamName).ListAuthMethods()
	if err == nil || len(authMethods) > 0 {
		err := fmt.Errorf("Team %s already exists", teamName)
		c.logger.Error("create-team.existing-team-error", err,
			lager.Data{
				"team-name":         teamName,
				"auth-methods-size": len(authMethods),
			})
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
