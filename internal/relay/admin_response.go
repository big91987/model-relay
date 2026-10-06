package relay

import (
	"bytes"
	"net/http"
)

type adminResponse struct {
	original http.ResponseWriter
	header   http.Header
	status   int
	body     bytes.Buffer
}

func newAdminResponse(writer http.ResponseWriter) *adminResponse {
	return &adminResponse{original: writer, header: writer.Header().Clone()}
}
func (response *adminResponse) Header() http.Header { return response.header }
func (response *adminResponse) WriteHeader(status int) {
	if response.status == 0 {
		response.status = status
	}
}
func (response *adminResponse) Write(data []byte) (int, error) {
	if response.status == 0 {
		response.status = 200
	}
	return response.body.Write(data)
}
func (response *adminResponse) Unwrap() http.ResponseWriter { return response.original }
func (response *adminResponse) flush() {
	for name, values := range response.header {
		response.original.Header()[name] = values
	}
	status := response.status
	if status == 0 {
		status = 200
	}
	response.original.WriteHeader(status)
	response.original.Write(response.body.Bytes())
}
