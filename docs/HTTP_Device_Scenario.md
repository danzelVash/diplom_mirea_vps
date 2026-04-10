# Device и Scenario API

Документ для локальной проверки двух уже реализованных сервисов: `device-service` и `scenario-service`.

## Как поднять

Из корня репозитория:

```bash
docker compose -f deploy/docker-compose.yaml up --build device-postgres scenario-postgres device-service scenario-service
```

После запуска сервисы будут доступны:

- `device-service HTTP` на `http://localhost:8082`
- `scenario-service HTTP` на `http://localhost:8084`

## Device Service

### Проверка доступности

```bash
curl http://localhost:8082/health
```

### Создать комнату

```bash
curl -X POST http://localhost:8082/api/v1/rooms \
  -H "Content-Type: application/json" \
  -d '{
    "edge_id": "edge-main",
    "name": "Гостиная",
    "floor": "1"
  }'
```

### Получить список комнат

```bash
curl "http://localhost:8082/api/v1/rooms?edge_id=edge-main"
```

### Создать устройство

```bash
curl -X POST http://localhost:8082/api/v1/devices \
  -H "Content-Type: application/json" \
  -d '{
    "edge_id": "edge-main",
    "room_id": "room_id_из_ответа",
    "name": "Основной свет",
    "device_type": "light",
    "entity_id": "light.living_room_main",
    "state": "off",
    "offline_capable": true
  }'
```

### Получить список устройств

```bash
curl "http://localhost:8082/api/v1/devices?edge_id=edge-main"
```

### Обновить состояние устройства

```bash
curl -X POST "http://localhost:8082/api/v1/devices/device_id_из_ответа/state?edge_id=edge-main" \
  -H "Content-Type: application/json" \
  -d '{
    "entity_id": "light.living_room_main",
    "state": "on"
  }'
```

### Выполнить команду устройству

```bash
curl -X POST http://localhost:8082/api/v1/devices/commands \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "device_id_из_ответа",
    "entity_id": "light.living_room_main",
    "target_state": "off",
    "source": "manual-test"
  }'
```

## Scenario Service

### Проверка доступности

```bash
curl http://localhost:8084/health
```

### Создать сценарий

```bash
curl -X POST http://localhost:8084/api/v1/scenarios \
  -H "Content-Type: application/json" \
  -d '{
    "edge_id": "edge-main",
    "name": "Включение света по движению",
    "enabled": true,
    "priority": 100,
    "offline_eligible": true,
    "triggers": [
      {
        "trigger_type": "event",
        "event_type": "motion_detected",
        "entity_id": "sensor.hall_motion",
        "expected_state": "motion"
      }
    ],
    "conditions": [
      {
        "condition_type": "equals",
        "field": "room_id",
        "expected_value": "hall-room"
      }
    ],
    "actions": [
      {
        "action_type": "device_command",
        "device_id": "device_id_из_ответа",
        "entity_id": "light.living_room_main",
        "target_state": "on"
      }
    ]
  }'
```

### Получить список сценариев

```bash
curl "http://localhost:8084/api/v1/scenarios?edge_id=edge-main"
```

### Получить offline-сценарии

```bash
curl "http://localhost:8084/api/v1/scenarios/offline?edge_id=edge-main"
```

### Оценить событие и выполнить действия

```bash
curl -X POST http://localhost:8084/api/v1/scenarios/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "edge_id": "edge-main",
    "room_id": "hall-room",
    "entity_id": "sensor.hall_motion",
    "event_type": "motion_detected",
    "state": "motion"
  }'
```

Если сценарий совпадет, `scenario-service` вызовет `device-service` по gRPC и переведет устройство в целевое состояние.
