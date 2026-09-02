//go:build windows

package hertz

import (
	hertzserver "github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/network/standard"
)

func platformTransportOptions() []config.Option {
	return []config.Option{hertzserver.WithTransport(standard.NewTransporter)}
}
