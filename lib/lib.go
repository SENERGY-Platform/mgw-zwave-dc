package lib

import (
	"context"
	"mgw-zwave-dc/lib/configuration"
	"mgw-zwave-dc/lib/connector"
)

func New(config configuration.Config, ctx context.Context) (result *connector.Connector, err error) {
	return connector.New(config, ctx)
}
