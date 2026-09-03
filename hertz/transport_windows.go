//go:build windows

package hertz

import (
	hertzserver "github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/network"
	"github.com/cloudwego/hertz/pkg/network/standard"

	internaltransport "goark.dev/arkhos/hertz/internal/transport"
)

func platformTransportOptions(tracker *internaltransport.Tracker) []config.Option {
	return []config.Option{hertzserver.WithTransport(func(options *config.Options) network.Transporter {
		return tracker.Wrap(standard.NewTransporter(options))
	})}
}
