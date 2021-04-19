package lib

import (
	"context"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/configuration"
	"github.com/SENERGY-Platform/mgw-zwave-dc/lib/connector"
)

func New(config configuration.Config, ctx context.Context) (result *connector.Connector, err error) {
	return connector.New(config, ctx)
}
