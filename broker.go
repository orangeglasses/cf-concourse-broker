package main

import (
	"context"
	"errors"

	"code.cloudfoundry.org/lager"
	"github.com/pivotal-cf/brokerapi"
)

type broker struct {
	services []brokerapi.Service
	logger   lager.Logger
	env      brokerConfig
}

func (b *broker) Services(context context.Context) []brokerapi.Service {
	return b.services
}

func (b *broker) Provision(context context.Context, instanceID string, details brokerapi.ProvisionDetails, asyncAllowed bool) (brokerapi.ProvisionedServiceSpec, error) {
	cfClient, err := cfNewClient(b.env)
	if err != nil {
		return brokerapi.ProvisionedServiceSpec{}, err
	}

	orgName, err := cfClient.getOrgNameBySpaceGuid(details.SpaceGUID)
	if err != nil {
		return brokerapi.ProvisionedServiceSpec{}, err
	}

	concourseClient := concourseNewClient(b.env, b.logger)
	return brokerapi.ProvisionedServiceSpec{}, concourseClient.CreateTeam(orgName)
}

func (b *broker) Deprovision(context context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {
	//get a CF Client to lookup orgName
	cfClient, err := cfNewClient(b.env)
	if err != nil {
		return brokerapi.DeprovisionServiceSpec{}, err
	}

	//lookup service instance to find space guid
	serviceInstance, err := cfClient.getServiceInstanceByGuid(details.ServiceID)
	if err != nil {
		return brokerapi.DeprovisionServiceSpec{}, err
	}

	//use space guid to lookup org name. Orgname is used as team name
	orgName, err := cfClient.getOrgNameBySpaceGuid(serviceInstance.SpaceGuid)
	if err != nil {
		return brokerapi.DeprovisionServiceSpec{}, err
	}

	concourseClient := concourseNewClient(b.env, b.logger)
	return brokerapi.DeprovisionServiceSpec{}, concourseClient.DeleteTeam(orgName)
}

func (b *broker) Bind(context context.Context, instanceID, bindingID string, details brokerapi.BindDetails) (brokerapi.Binding, error) {
	return brokerapi.Binding{}, errors.New("This service does not support bind")
}

func (b *broker) Unbind(context context.Context, instanceID, bindingID string, details brokerapi.UnbindDetails) error {
	return errors.New("This service does not support bind")
}

func (b *broker) Update(context context.Context, instanceID string, details brokerapi.UpdateDetails, asyncAllowed bool) (brokerapi.UpdateServiceSpec, error) {
	return brokerapi.UpdateServiceSpec{}, nil
}

func (b *broker) LastOperation(context context.Context, instanceID, operationData string) (brokerapi.LastOperation, error) {
	return brokerapi.LastOperation{}, nil
}
