package restful

import (
	"encoding/json"
	"net/http"
	"reflect"
)

type PaginateResult[T any] struct {
	TotalItems int64 `json:"totalItems"`
	Results    []T   `json:"results"`
}

type DefaultResponse struct {
	Data   any `json:"data"`
	Status int `json:"status"`
	Errors any `json:"errors"`
}

type IResponder interface {
	Respond(data any, err error)
}

type Responder struct {
	writer http.ResponseWriter
	method string
}

func NewResponder(writer http.ResponseWriter, method string) *Responder {
	return &Responder{writer: writer, method: method}
}

func (r Responder) Respond(data any, err error) {
	statusCode, errorObj := getStatusCode(r.method, data, err)

	var resp any
	switch v := data.(type) {
	case DefaultResponse:
		resp = DefaultResponse{Data: v.Data, Status: v.Status, Errors: errorObj}
		return
	default:
		if isNil(data) {
			data = nil
		}

		resp = DefaultResponse{Data: data, Status: statusCode, Errors: errorObj}
	}

	r.writer.Header().Set("Content-Type", "application/json")

	r.writer.WriteHeader(statusCode)

	json.NewEncoder(r.writer).Encode(resp)
}

func getStatusCode(method string, data any, err error) (statusCode int, errResp any) {
	if err == nil {
		return handleSuccess(method, data)
	}

	if !isNil(data) {
		return http.StatusPartialContent, createErrorResponse(err)
	}

	if e, ok := err.(statusCodeResponder); ok {
		return e.StatusCode(), createErrorResponse(err)
	}

	return http.StatusInternalServerError, createErrorResponse(err)
}

func handleSuccess(method string, data any) (statusCode int, err any) {
	switch method {
	case http.MethodPost:
		if data != nil {
			return http.StatusCreated, nil
		}

		return http.StatusAccepted, nil
	case http.MethodDelete:
		return http.StatusNoContent, nil
	default:
		return http.StatusOK, nil
	}
}

func createErrorResponse(err error) string {
	return err.Error()
}

type statusCodeResponder interface {
	StatusCode() int
}

func isNil(i any) bool {
	if i == nil {
		return true
	}

	v := reflect.ValueOf(i)

	return v.Kind() == reflect.Ptr && v.IsNil()
}
