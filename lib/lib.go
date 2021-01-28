package lib

import (
	"context"
	"zwave2mqtt-connector/lib/configuration"
	"zwave2mqtt-connector/lib/connector"
)

func New(config configuration.Config, ctx context.Context) (result *connector.Connector, err error) {
	return connector.New(config, ctx)
}
