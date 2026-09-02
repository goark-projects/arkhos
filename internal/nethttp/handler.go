package nethttp

import (
	"net/http"

	arkartanethttp "goark.dev/arkarta/servlet/nethttp"

	"goark.dev/arkarta/servlet"
	servletcontainer "goark.dev/arkarta/servlet/container"
)

// Handler 将 Arkarta Servlet Handler 适配为标准库 http.Handler。
func Handler(handler servlet.Handler, options ...arkartanethttp.Option) http.Handler {
	return arkartanethttp.HandlerWithOptions(handler, options...)
}

// ApplicationHandler 将已部署应用适配为标准库 http.Handler。
func ApplicationHandler(application servletcontainer.Application, requestOptions ...servlet.RequestOption) http.Handler {
	if application == nil || application.WebApp() == nil {
		return unavailableHandler()
	}
	return Handler(
		application.Handler(),
		arkartanethttp.WithRequestContextPath(application.WebApp().ContextPath()),
		arkartanethttp.WithRequestOptions(requestOptions...),
	)
}

func unavailableHandler() http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		http.Error(writer, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
	})
}
