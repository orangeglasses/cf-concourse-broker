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
	cfDetails, err := cfClient.GetProvisionDetails(details.SpaceGUID)
	cfDetails.SpaceGUID = details.SpaceGUID
	if err != nil {
		return brokerapi.ProvisionedServiceSpec{}, err
	}
	concourseClient := concourseNewClient(b.env, b.logger)
	err = concourseClient.CreateTeam(cfDetails)
	if err != nil {
		return brokerapi.ProvisionedServiceSpec{}, err
	}
	return brokerapi.ProvisionedServiceSpec{}, nil
}

func (b *broker) Deprovision(context context.Context, instanceID string, details brokerapi.DeprovisionDetails, asyncAllowed bool) (brokerapi.DeprovisionServiceSpec, error) {
	return brokerapi.DeprovisionServiceSpec{}, nil
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
