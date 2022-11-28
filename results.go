package rez

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type NotFound[V any] struct {
	Result V `json:"result,omitempty"`
}

var _ error = &NotFound[string]{}
var _ HasStatus = &NotFound[string]{}

func (err NotFound[V]) HTTPStatus() int {
	return http.StatusNotFound
}
func (err NotFound[V]) HTTPStatuses() []int {
	return []int{http.StatusNotFound}
}
func (err NotFound[V]) Error() string {
	msg := fmt.Sprintf("%v", err.Result)
	if msg == "" {
		msg = http.StatusText(err.HTTPStatus())
	}
	return msg
}

type OK[V any] struct {
	Result V `json:"result,omitempty"`
}

var _ error = &OK[string]{}
var _ HasStatus = &OK[string]{}

func (err OK[V]) HTTPStatus() int {
	return http.StatusOK
}
func (err OK[V]) HTTPStatuses() []int {
	return []int{http.StatusOK}
}
func (err OK[V]) Error() string {
	msg := fmt.Sprintf("%v", err.Result)
	if msg == "" {
		msg = http.StatusText(err.HTTPStatus())
	}
	return msg
}

type Created[V any] struct {
	Result V `json:"result,omitempty"`
}

var _ error = &Created[string]{}
var _ HasStatus = &Created[string]{}

func (err Created[V]) HTTPStatus() int {
	return http.StatusCreated
}
func (err Created[V]) HTTPStatuses() []int {
	return []int{http.StatusCreated}
}
func (err Created[V]) Error() string {
	msg := fmt.Sprintf("%v", err.Result)
	if msg == "" {
		msg = http.StatusText(err.HTTPStatus())
	}
	return msg
}

type Accepted[V any] struct {
	Result V `json:"result,omitempty"`
}

var _ error = &Accepted[string]{}
var _ HasStatus = &Accepted[string]{}

func (err Accepted[V]) HTTPStatus() int {
	return http.StatusAccepted
}
func (err Accepted[V]) HTTPStatuses() []int {
	return []int{http.StatusAccepted}
}
func (err Accepted[V]) Error() string {
	msg := fmt.Sprintf("%v", err.Result)
	if msg == "" {
		msg = http.StatusText(err.HTTPStatus())
	}
	return msg
}

type BadRequest[V any] struct {
	Result V `json:"result,omitempty"`
}

var _ error = &BadRequest[string]{}
var _ HasStatus = &BadRequest[string]{}

func (err BadRequest[V]) HTTPStatus() int {
	return http.StatusBadRequest
}
func (err BadRequest[V]) HTTPStatuses() []int {
	return []int{http.StatusBadRequest}
}
func (err BadRequest[V]) Error() string {
	msg := fmt.Sprintf("%v", err.Result)
	if msg == "" {
		msg = http.StatusText(err.HTTPStatus())
	}
	return msg
}

type Unauthorized[V any] struct {
	Result V `json:"result,omitempty"`
}

var _ error = &Unauthorized[string]{}
var _ HasStatus = &Unauthorized[string]{}

func (err Unauthorized[V]) HTTPStatus() int {
	return http.StatusUnauthorized
}
func (err Unauthorized[V]) HTTPStatuses() []int {
	return []int{http.StatusUnauthorized}
}
func (err Unauthorized[V]) Error() string {
	msg := fmt.Sprintf("%v", err.Result)
	if msg == "" {
		msg = http.StatusText(err.HTTPStatus())
	}
	return msg
}

type Forbidden[V any] struct {
	Result V `json:"result,omitempty"`
}

var _ error = &Forbidden[string]{}
var _ HasStatus = &Forbidden[string]{}

func (err Forbidden[V]) HTTPStatus() int {
	return http.StatusForbidden
}
func (err Forbidden[V]) HTTPStatuses() []int {
	return []int{http.StatusForbidden}
}
func (err Forbidden[V]) Error() string {
	msg := fmt.Sprintf("%v", err.Result)
	if msg == "" {
		msg = http.StatusText(err.HTTPStatus())
	}
	return msg
}

type Conflict[V any] struct {
	Result V `json:"result,omitempty"`
}

var _ error = &Conflict[string]{}
var _ HasStatus = &Conflict[string]{}

func (err Conflict[V]) HTTPStatus() int {
	return http.StatusConflict
}
func (err Conflict[V]) HTTPStatuses() []int {
	return []int{http.StatusConflict}
}
func (err Conflict[V]) Error() string {
	msg := fmt.Sprintf("%v", err.Result)
	if msg == "" {
		msg = http.StatusText(err.HTTPStatus())
	}
	return msg
}

type TooManyRequests[V any] struct {
	Result V `json:"result,omitempty"`
}

var _ error = &TooManyRequests[string]{}
var _ HasStatus = &TooManyRequests[string]{}

func (err TooManyRequests[V]) HTTPStatus() int {
	return http.StatusTooManyRequests
}
func (err TooManyRequests[V]) HTTPStatuses() []int {
	return []int{http.StatusTooManyRequests}
}
func (err TooManyRequests[V]) Error() string {
	msg := fmt.Sprintf("%v", err.Result)
	if msg == "" {
		msg = http.StatusText(err.HTTPStatus())
	}
	return msg
}

type InternalServerError[V any] struct {
	Result V `json:"result,omitempty"`
}

var _ error = &InternalServerError[string]{}
var _ HasStatus = &InternalServerError[string]{}

func (err InternalServerError[V]) HTTPStatus() int {
	return http.StatusInternalServerError
}
func (err InternalServerError[V]) HTTPStatuses() []int {
	return []int{http.StatusInternalServerError}
}
func (err InternalServerError[V]) Error() string {
	msg := fmt.Sprintf("%v", err.Result)
	if msg == "" {
		msg = http.StatusText(err.HTTPStatus())
	}
	return msg
}

type NotImplemented[V any] struct {
	Result V `json:"result,omitempty"`
}

var _ error = &NotImplemented[string]{}
var _ HasStatus = &NotImplemented[string]{}

func (err NotImplemented[V]) HTTPStatus() int {
	return http.StatusNotImplemented
}
func (err NotImplemented[V]) HTTPStatuses() []int {
	return []int{http.StatusNotImplemented}
}
func (err NotImplemented[V]) Error() string {
	msg := fmt.Sprintf("%v", err.Result)
	if msg == "" {
		msg = http.StatusText(err.HTTPStatus())
	}
	return msg
}

type ServiceUnavailable[V any] struct {
	Result V `json:"result,omitempty"`
}

var _ error = &ServiceUnavailable[string]{}
var _ HasStatus = &ServiceUnavailable[string]{}

func (err ServiceUnavailable[V]) HTTPStatus() int {
	return http.StatusServiceUnavailable
}
func (err ServiceUnavailable[V]) HTTPStatuses() []int {
	return []int{http.StatusServiceUnavailable}
}
func (err ServiceUnavailable[V]) Error() string {
	msg := fmt.Sprintf("%v", err.Result)
	if msg == "" {
		msg = http.StatusText(err.HTTPStatus())
	}
	return msg
}

type Result[V any] struct {
	Status int
	Value  V
}

var _ json.Marshaler = &Result[string]{}
var _ json.Unmarshaler = &Result[string]{}
var _ HasStatus = &Result[string]{}

func (se Result[V]) HTTPStatus() int {
	return se.Status
}
func (se Result[V]) HTTPStatuses() []int {
	return []int{se.Status}
}
func (se Result[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(se.Value)
}
func (se *Result[V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &se.Value)
}
