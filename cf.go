package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry-community/go-cfclient"
)

type cfDetails struct {
	OrgGUID   string
	OrgName   string
	SpaceGUID string
	SpaceName string
}

type IcfClient interface {
	GetProvisionDetails(spaceGUID string) (cfDetails, error)
	GetDeprovisionDetails(serviceGUID string) (cfDetails, error)
}

func cfNewClient(config brokerConfig) (IcfClient, error) {
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

func (c *cfClient) GetProvisionDetails(spaceGUID string) (cfDetails, error) {
	requestURI := fmt.Sprintf("/v2/spaces/%s", spaceGUID)
	orgName, err := c.getOrgName(requestURI)
	if err != nil {
		return cfDetails{}, err
	}
	return cfDetails{OrgName: orgName}, nil
}

func (c *cfClient) GetDeprovisionDetails(serviceGUID string) (cfDetails, error) {
	serviceInstance, err := c.client.ServiceInstanceByGuid(serviceGUID)
	if err != nil {
		return cfDetails{}, err
	}
	orgName, err := c.getOrgName(serviceInstance.SpaceUrl)
	if err != nil {
		return cfDetails{}, err
	}
	return cfDetails{OrgName: orgName}, nil
}

func (c *cfClient) getOrgName(requestUrl string) (string, error) {
	var spaceResp cfclient.SpaceResource
	r := c.client.NewRequest("GET", requestUrl)
	resp, err := c.client.DoRequest(r)
	if err != nil {
		return "", fmt.Errorf("Error requesting spaces %v", err)
	}
	resBody, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("Error reading space request %v", err)
	}
	err = json.Unmarshal(resBody, &spaceResp)
	if err != nil {
		return "", fmt.Errorf("Error unmarshalling space %v", err)
	}
	var orgResp cfclient.OrgResource
	r = c.client.NewRequest("GET", spaceResp.Entity.OrgURL)
	resp, err = c.client.DoRequest(r)
	if err != nil {
		return "", fmt.Errorf("Error requesting orgs %v", err)
	}
	resBody, err = ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", fmt.Errorf("Error reading org request %v", err)
	}
	err = json.Unmarshal(resBody, &orgResp)
	if err != nil {
		return "", fmt.Errorf("Error unmarshalling org %v", err)
	}
	return orgResp.Entity.Name, nil
}
