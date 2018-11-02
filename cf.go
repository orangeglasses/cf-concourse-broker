package main

import (
	"github.com/cloudfoundry-community/go-cfclient"
)

func cfNewClient(config brokerConfig) (*cfClient, error) {
	cfConfig := &cfclient.Config{
		ClientID:     config.ClientID,
		ClientSecret: config.ClientSecret,
		ApiAddress:   config.CFURL,
	}
	client, err := cfclient.NewClient(cfConfig)
	if err != nil {
		return nil, err
	}
	return &cfClient{client: client}, nil
}

type cfClient struct {
	client *cfclient.Client
}

func (c *cfClient) getServiceInstanceByGuid(siGUID string) (cfclient.ServiceInstance, error) {
	return c.client.ServiceInstanceByGuid(siGUID)
}

func (c *cfClient) getOrgNameBySpaceGuid(spaceGUID string) (string, error) {
	space, err := c.client.GetSpaceByGuid(spaceGUID)
	if err != nil {
		return "", err
	}

	org, _ := space.Org()
	return org.Name, nil
}
