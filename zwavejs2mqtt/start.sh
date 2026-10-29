# update config with environment variables
#
# settings.json lives in the store volume and holds the Z-Wave security keys.
# It is only replaced once jq has produced a complete result. A failed jq run
# must not truncate it, the keys cannot be restored from anywhere else.

settings=/usr/src/app/store/settings.json
tmp="${settings}.tmp"

# seed the defaults if the volume has no settings yet
if [ ! -s "${settings}" ]; then
  cp /usr/src/app/settings.default.json "${settings}"
fi

# use argjson to set port, qos and log flags as number or boolean
if jq \
  --argjson port "${MQTT_BROKER_PORT}" \
  --argjson qos "${MQTT_QOS}" \
  --argjson gatewayLogEnabled "${GATEWAY_LOG_ENABLED}" \
  --argjson zwaveLogEnabled "${ZWAVE_LOG_ENABLED}" \
  '.mqtt.host = env.MQTT_BROKER_HOST
  | .mqtt.port = $port
  | .mqtt.qos = $qos
  | .mqtt.prefix = env.MQTT_TOPIC_PREFIX
  | .gateway.logEnabled = $gatewayLogEnabled
  | .gateway.logLevel = env.GATEWAY_LOG_LEVEL
  | .zwave.logEnabled = $zwaveLogEnabled
  | .zwave.logLevel = env.ZWAVE_LOG_LEVEL' \
  "${settings}" > "${tmp}" && [ -s "${tmp}" ]; then
  mv "${tmp}" "${settings}"
else
  echo "could not apply environment to ${settings}, starting with the file unchanged" >&2
  rm -f "${tmp}"
fi

# print the settings without key material
jq 'del(.zwave.securityKeys, .zwave.securityKeysLongRange, .zwave.networkKey, .zniffer.securityKeys, .zniffer.securityKeysLongRange)' "${settings}"

exec node server/bin/www
