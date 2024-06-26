package exporter

import "net/http"

type Exporter interface {
	Wrapper(handlerName string) http.HandlerFunc
}
