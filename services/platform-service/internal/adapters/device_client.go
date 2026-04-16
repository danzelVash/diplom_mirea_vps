package adapters

import (
	"context"

	"google.golang.org/grpc"
	devicev1 "platform-service/pkg/pb/device/v1"
)

type DeviceClient struct {
	Server devicev1.DeviceServiceServer
}

func (c DeviceClient) ListRooms(ctx context.Context, in *devicev1.ListRoomsRequest, _ ...grpc.CallOption) (*devicev1.ListRoomsResponse, error) {
	return c.Server.ListRooms(ctx, in)
}

func (c DeviceClient) ListDevices(ctx context.Context, in *devicev1.ListDevicesRequest, _ ...grpc.CallOption) (*devicev1.ListDevicesResponse, error) {
	return c.Server.ListDevices(ctx, in)
}

func (c DeviceClient) GetDevice(ctx context.Context, in *devicev1.GetDeviceRequest, _ ...grpc.CallOption) (*devicev1.GetDeviceResponse, error) {
	return c.Server.GetDevice(ctx, in)
}

func (c DeviceClient) SyncInventory(ctx context.Context, in *devicev1.SyncInventoryRequest, _ ...grpc.CallOption) (*devicev1.SyncInventoryResponse, error) {
	return c.Server.SyncInventory(ctx, in)
}

func (c DeviceClient) UpsertDeviceState(ctx context.Context, in *devicev1.UpsertDeviceStateRequest, _ ...grpc.CallOption) (*devicev1.UpsertDeviceStateResponse, error) {
	return c.Server.UpsertDeviceState(ctx, in)
}

func (c DeviceClient) ExecuteCommand(ctx context.Context, in *devicev1.ExecuteCommandRequest, _ ...grpc.CallOption) (*devicev1.ExecuteCommandResponse, error) {
	return c.Server.ExecuteCommand(ctx, in)
}
