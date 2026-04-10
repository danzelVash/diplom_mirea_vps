package v1

import (
	"context"

	"device-service/internal/service"
	devicev1 "device-service/pkg/pb/device/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Implementation struct {
	devicev1.UnimplementedDeviceServiceServer
	service *service.Service
}

func New(service *service.Service) *Implementation {
	return &Implementation{service: service}
}

func (i *Implementation) ListRooms(ctx context.Context, req *devicev1.ListRoomsRequest) (*devicev1.ListRoomsResponse, error) {
	rooms, err := i.service.ListRooms(ctx, req.GetEdgeId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list rooms: %v", err)
	}

	response := &devicev1.ListRoomsResponse{
		Rooms: make([]*devicev1.Room, 0, len(rooms)),
	}
	for _, room := range rooms {
		response.Rooms = append(response.Rooms, room.ToProtoRoom())
	}
	return response, nil
}

func (i *Implementation) ListDevices(ctx context.Context, req *devicev1.ListDevicesRequest) (*devicev1.ListDevicesResponse, error) {
	devices, err := i.service.ListDevices(ctx, service.ListDevicesFilter{
		EdgeID: req.GetEdgeId(),
		RoomID: req.GetRoomId(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list devices: %v", err)
	}

	response := &devicev1.ListDevicesResponse{
		Devices: make([]*devicev1.Device, 0, len(devices)),
	}
	for _, device := range devices {
		response.Devices = append(response.Devices, device.ToProto())
	}
	return response, nil
}

func (i *Implementation) GetDevice(ctx context.Context, req *devicev1.GetDeviceRequest) (*devicev1.GetDeviceResponse, error) {
	device, err := i.service.GetDevice(ctx, req.GetDeviceId())
	if err != nil {
		if err == service.ErrNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "get device: %v", err)
	}

	return &devicev1.GetDeviceResponse{Device: device.ToProto()}, nil
}

func (i *Implementation) SyncInventory(ctx context.Context, req *devicev1.SyncInventoryRequest) (*devicev1.SyncInventoryResponse, error) {
	syncID, err := i.service.SyncInventory(ctx, req)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "sync inventory: %v", err)
	}

	return &devicev1.SyncInventoryResponse{
		SyncId: syncID,
		Status: "synced",
	}, nil
}

func (i *Implementation) UpsertDeviceState(ctx context.Context, req *devicev1.UpsertDeviceStateRequest) (*devicev1.UpsertDeviceStateResponse, error) {
	if err := i.service.UpsertDeviceState(ctx, req); err != nil {
		if err == service.ErrNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "upsert device state: %v", err)
	}

	return &devicev1.UpsertDeviceStateResponse{Status: "updated"}, nil
}

func (i *Implementation) ExecuteCommand(ctx context.Context, req *devicev1.ExecuteCommandRequest) (*devicev1.ExecuteCommandResponse, error) {
	commandID, err := i.service.ExecuteCommand(ctx, req)
	if err != nil {
		if err == service.ErrNotFound {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Errorf(codes.Internal, "execute command: %v", err)
	}

	return &devicev1.ExecuteCommandResponse{
		CommandId: commandID,
		Status:    "executed",
	}, nil
}
