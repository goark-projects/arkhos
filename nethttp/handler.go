package nethttp

import (
	"net/http"

	arkartanethttp "goark.dev/arkarta/servlet/nethttp"

	"goark.dev/arkarta/servlet"
)

// Handler 将 Arkarta Servlet Handler 适配为标准库 http.Handler。
func Handler(handler servlet.Handler) http.Handler {
	return HandlerWithOptions(handler)
}

// HandlerWithOptions 按配置将 Arkarta Servlet Handler 适配为标准库 http.Handler。
func HandlerWithOptions(handler servlet.Handler, options ...arkartanethttp.Option) http.Handler {
	return arkartanethttp.HandlerWithOptions(handler, options...)
}
