package restful

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-base/pkg/common"
	"io"
	"net/http"
	"reflect"
	"strings"
)

const (
	defaultMaxMemory = 32 << 20 // 32 MB
)

var (
	errNoFileFound    = errors.New("no files were bounded")
	errNonPointerBind = errors.New("bind error, cannot bind to a non pointer type")
	errNonSliceBind   = errors.New("bind error: input is not a pointer to a byte slice")
)

type IRequest interface {
	Context() context.Context
	Query(string) string
	PathParam(string) string
	Bind(any) error
	HostName() string
	Params(string) []string
}

type Request struct {
	req *http.Request
}

func NewRequest(r *http.Request) *Request {
	return &Request{
		req: r,
	}
}

func (r *Request) Query(key string) string {
	return r.req.URL.Query().Get(key)
}

func (r *Request) Context() context.Context {
	return r.req.Context()
}

func (r *Request) PathParam(key string) string {
	params, ok := r.Context().Value(common.ReqParams).(map[string]string)

	if !ok || params == nil {
		return ""
	}

	return params[key]
}

func (r *Request) Bind(i any) error {
	v := r.req.Header.Get("Content-Type")
	contentType := strings.Split(v, ";")[0]

	switch contentType {
	case "application/json":
		body, err := r.body()
		if err != nil {
			return err
		}

		return json.Unmarshal(body, &i)
	case "multipart/form-data":
		return r.bindForm(i, true)
	// case "application/x-www-form-urlencoded":
	// 	return r.bindFormURLEncoded(i)
	case "binary/octet-stream":
		return r.bindBinary(i)
	}

	return nil
}

func (r *Request) HostName() string {
	proto := r.req.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
	}

	return fmt.Sprintf("%s://%s", proto, r.req.Host)
}

func (r *Request) Params(key string) []string {
	values := r.req.URL.Query()[key]

	var result []string

	for _, value := range values {
		result = append(result, strings.Split(value, ",")...)
	}

	return result
}

func (r *Request) body() ([]byte, error) {
	bodyBytes, err := io.ReadAll(r.req.Body)
	if err != nil {
		return nil, err
	}

	r.req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	return bodyBytes, nil
}

// bindBinary handles binding for binary/octet-stream content type.
func (r *Request) bindBinary(raw any) error {
	// Ensure raw is a pointer to a byte slice
	byteSlicePtr, ok := raw.(*[]byte)
	if !ok {
		return fmt.Errorf("%w: %v", errNonSliceBind, raw)
	}

	body, err := r.body()
	if err != nil {
		return fmt.Errorf("failed to read request body: %w", err)
	}

	// Assign the body to the provided slice
	*byteSlicePtr = body

	return nil
}

func (r *Request) bindForm(ptr any, isMultipart bool) error {
	ptrVal := reflect.ValueOf(ptr)
	if ptrVal.Kind() != reflect.Ptr {
		return errNonPointerBind
	}

	ptrVal = ptrVal.Elem()

	var fd formData

	if isMultipart {
		if err := r.req.ParseMultipartForm(defaultMaxMemory); err != nil {
			return err
		}

		fd = formData{files: r.req.MultipartForm.File, fields: r.req.MultipartForm.Value}
	} else {
		if err := r.req.ParseForm(); err != nil {
			return err
		}

		fd = formData{fields: r.req.Form}
	}

	ok, err := fd.mapStruct(ptrVal, nil)
	if err != nil {
		return err
	}

	if !ok {
		if isMultipart {
			return errNoFileFound
		}

		return errFieldsNotSet
	}

	return nil
}
