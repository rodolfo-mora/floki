package exporter

import "net/http"

type Exporter interface {
	Wrapper(handlerName string, hand func(w http.ResponseWriter, r *http.Request)) http.HandlerFunc
	Export() http.Handler
}
