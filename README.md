## Auth
The 'auth_endpoint' config field can/should be left empty if this component is used in a MGW with API-Proxy

## zwavejs2mqtt configurations
- Payload type: Entire Z-Wave value Object
- Send Zwave events: (optional to enable live device register)

## command acknowledgement
Only with `ZWAVE_CONTROLLER=zwavejs2mqtt`: a set command is answered after zwave-js-ui responded to the `writeValue` call. A failed write or no response within `COMMAND_ACK_TIMEOUT` is reported on `error/command/<command_id>` instead.
`COMMAND_ACK_TIMEOUT` (`command_ack_timeout`) is a Go duration. The default `30s` comes from the image's `config.json`; a mounted config without the field disables the wait. `-` or a zero duration disable it as well (in the config file also `""`; an empty environment variable is ignored), the command is then answered right after it was sent. An invalid value stops the connector at start, with either controller; a valid value has no effect with the zwave2mqtt controller.
A command sent right after the connector (re)connected to the broker can miss its response, because the subscription on `<api topic>/writeValue` is set up asynchronously on connect; it then ends with the timeout error although the write may have succeeded.

## zwavejs2mqtt image
`zwavejs2mqtt/` builds `ghcr.io/senergy-platform/zwavejs2mqtt`, the zwave-js-ui image used by the mgw module. Its start script applies the MQTT and log settings from the environment to `settings.json` in the store volume.
The image has its own version line: a tag `zwavejs2mqtt/vX.Y.Z` publishes the image as `vX.Y.Z`, plain `vX.Y.Z` tags release mgw-zwave-dc.

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