## Auth
The 'auth_endpoint' config field can/should be left empty if this component is used in a MGW with API-Proxy

## zwavejs2mqtt configurations
- Payload type: Entire Z-Wave value Object
- Send Zwave events: (optional to enable live device register)

## deployment example
```
docker run -d --name zwavejs2mqtt -p 8091:8091 -p 3000:3000 --device=/dev/serial/by-id/usb-0658_0200-if00:/dev/zwave \
--mount source=zwavejs2mqtt,target=/usr/src/app/store zwavejs/zwavejs2mqtt:latest

docker run -d --name mgw-zwave-dc \
  -e ZWAVE_MQTT_BROKER=tcp://localhost:1883 \
  -e MGW_MQTT_BROKER=tcp://localhost:1883 \
  -e ZWAVE_CONTROLLER=zwavejs2mqtt \
  -e ZWAVE_MQTT_API_TOPIC=zwave/_CLIENTS/ZWAVE_GATEWAY-Zwavejs2Mqtt/api \
  -e ZWAVE_NETWORK_EVENTS_TOPIC=zwave/_EVENTS/ZWAVE_GATEWAY-Zwavejs2Mqtt/node \
  -e ZWAVE_MQTT_DEVICE_STATE_TOPIC=zwave/# \
  -e DEBUG=true \
  mgw-zwave-dc:test
```