//go:build !windows

package hertz

import "github.com/cloudwego/hertz/pkg/common/config"

func platformTransportOptions() []config.Option {
	return nil
}
