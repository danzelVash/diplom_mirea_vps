package v1

import (
	devicev1 "device-service/pkg/pb/device/v1"
	edgebridgev1 "edge-bridge-service/pkg/pb/edge_bridge/v1"
	scenariov1 "scenario-service/pkg/pb/scenario/v1"
)

type Implementation struct {
	edgebridgev1.UnimplementedEdgeBridgeServiceServer
	device   externalDeviceClient
	scenario externalScenarioClient
}

type externalDeviceClient interface {
	devicev1.DeviceServiceClient
}

type externalScenarioClient interface {
	scenariov1.ScenarioServiceClient
}

func New(device externalDeviceClient, scenario externalScenarioClient) *Implementation {
	return &Implementation{
		device:   device,
		scenario: scenario,
	}
}
