# Proto-контракты VPS-контура

## Назначение

Документ фиксирует состав gRPC-контрактов между сервисами VPS-контура и показывает, какие сервисы выступают сервером, а какие клиентом.

## Общий принцип

- у каждого сервиса есть собственный `.proto`-контракт в каталоге `api/`;
- generated `pb`-код лежит в `pkg/pb/...`;
- серверная реализация зарегистрирована в `internal/app/.../v1/service.go`;
- текущая реализация методов является заглушкой через `Unimplemented...Server`;
- typed gRPC-клиенты уже инициализируются на старте там, где сервис зависит от других сервисов.

## Контракты по сервисам

### `api-gateway`

Файл:

- [gateway.proto](/Users/daniilbulykin/Desktop/diplom/diplom_mirea_vps/services/api-gateway/api/api_gateway/v1/gateway.proto)

Сервис:

- `GatewayService`

RPC:

- `GetDashboard`
- `ExecuteDeviceCommand`
- `SaveScenario`
- `ParseVoiceCommand`
- `AnalyzeFrame`

Клиентские зависимости:

- `edge-bridge-service`
- `device-service`
- `context-service`
- `scenario-service`
- `voice-service`
- `vision-service`
- `notification-service`

### `edge-bridge-service`

Файл:

- [edge_bridge.proto](/Users/daniilbulykin/Desktop/diplom/diplom_mirea_vps/services/edge-bridge-service/api/edge_bridge/v1/edge_bridge.proto)

Сервис:

- `EdgeBridgeService`

RPC:

- `RegisterEdge`
- `SyncInventory`
- `PublishEvent`
- `PollCommands`
- `GetOfflineScenarios`

Клиентские зависимости:

- `device-service`
- `scenario-service`

### `device-service`

Файл:

- [device.proto](/Users/daniilbulykin/Desktop/diplom/diplom_mirea_vps/services/device-service/api/device/v1/device.proto)

Сервис:

- `DeviceService`

RPC:

- `ListRooms`
- `ListDevices`
- `GetDevice`
- `SyncInventory`
- `UpsertDeviceState`
- `ExecuteCommand`

Клиентские зависимости:

- отсутствуют

### `context-service`

Файл:

- [context.proto](/Users/daniilbulykin/Desktop/diplom/diplom_mirea_vps/services/context-service/api/context/v1/context.proto)

Сервис:

- `ContextService`

RPC:

- `GetHomeContext`
- `GetRoomContext`
- `IngestSignal`
- `GetPresenceContext`

Клиентские зависимости:

- `device-service`
- `voice-service`
- `vision-service`

### `scenario-service`

Файл:

- [scenario.proto](/Users/daniilbulykin/Desktop/diplom/diplom_mirea_vps/services/scenario-service/api/scenario/v1/scenario.proto)

Сервис:

- `ScenarioService`

RPC:

- `ListScenarios`
- `GetScenario`
- `SaveScenario`
- `EvaluateEvent`
- `GetOfflineScenarios`

Клиентские зависимости:

- `device-service`
- `context-service`
- `notification-service`

### `voice-service`

Файл:

- [voice.proto](/Users/daniilbulykin/Desktop/diplom/diplom_mirea_vps/services/voice-service/api/voice/v1/voice.proto)

Сервис:

- `VoiceService`

RPC:

- `ParseVoiceCommand`
- `MatchOfflinePhrase`

Клиентские зависимости:

- отсутствуют

### `vision-service`

Файл:

- [vision.proto](/Users/daniilbulykin/Desktop/diplom/diplom_mirea_vps/services/vision-service/api/vision/v1/vision.proto)

Сервис:

- `VisionService`

RPC:

- `AnalyzeFrame`
- `DetectAccessEvent`

Клиентские зависимости:

- отсутствуют

### `notification-service`

Файл:

- [notification.proto](/Users/daniilbulykin/Desktop/diplom/diplom_mirea_vps/services/notification-service/api/notification/v1/notification.proto)

Сервис:

- `NotificationService`

RPC:

- `SendNotification`
- `CreateIncident`
- `ListIncidents`

Клиентские зависимости:

- отсутствуют

## Карта связности

```text
api-gateway
  -> edge-bridge-service
  -> device-service
  -> context-service
  -> scenario-service
  -> voice-service
  -> vision-service
  -> notification-service

edge-bridge-service
  -> device-service
  -> scenario-service

context-service
  -> device-service
  -> voice-service
  -> vision-service

scenario-service
  -> device-service
  -> context-service
  -> notification-service
```

## Текущее состояние реализации

- `.proto`-контракты описаны и разложены по сервисам;
- `pb.go` и `grpc.pb.go` сгенерированы;
- gRPC server wiring поднят во всех сервисах;
- typed client wiring добавлен в сервисы-координаторы;
- все RPC-методы пока возвращают `Unimplemented`, что соответствует текущему этапу.
