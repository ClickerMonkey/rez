package rez

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/ClickerMonkey/rez/api"
)

// A contract for a result which returns a specific HTTP status.
type validResult interface {
	json.Marshaler
	json.Unmarshaler
	HasStatus
	api.HasSchemaType
	api.HasName
}

// A contract for a result that is considered an error.
type invalidResult interface {
	error
	validResult
}

// A 200 response.
type OK[V any] struct {
	Result V
}

func NewOK[V any](result V) *OK[V] {
	return &OK[V]{result}
}

var _ validResult = &OK[string]{}

func (err OK[V]) HTTPStatus() int {
	return http.StatusOK
}
func (err OK[V]) HTTPStatuses() []int {
	return []int{http.StatusOK}
}
func (err OK[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Result)
}
func (err *OK[V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &err.Result)
}
func (err OK[V]) APISchemaType() any {
	return err.Result
}
func (err OK[V]) APIName() string {
	return getResultAPIName(err.Result, "OK")
}

// A 201 response.
type Created[V any] struct {
	Result V
}

func NewCreated[V any](result V) *Created[V] {
	return &Created[V]{result}
}

var _ validResult = &Created[string]{}

func (err Created[V]) HTTPStatus() int {
	return http.StatusCreated
}
func (err Created[V]) HTTPStatuses() []int {
	return []int{http.StatusCreated}
}
func (err Created[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Result)
}
func (err *Created[V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &err.Result)
}
func (err Created[V]) APISchemaType() any {
	return err.Result
}
func (err Created[V]) APIName() string {
	return getResultAPIName(err.Result, "Created")
}

// A 202 response.
type Accepted[V any] struct {
	Result V
}

func NewAccepted[V any](result V) *Accepted[V] {
	return &Accepted[V]{result}
}

var _ validResult = &Accepted[string]{}

func (err Accepted[V]) HTTPStatus() int {
	return http.StatusAccepted
}
func (err Accepted[V]) HTTPStatuses() []int {
	return []int{http.StatusAccepted}
}
func (err Accepted[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Result)
}
func (err *Accepted[V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &err.Result)
}
func (err Accepted[V]) APISchemaType() any {
	return err.Result
}
func (err Accepted[V]) APIName() string {
	return getResultAPIName(err.Result, "Accepted")
}

// A 301 response.
type Moved[V any] struct {
	Result V
}

func NewMoved[V any](result V) *Moved[V] {
	return &Moved[V]{result}
}

var _ validResult = &Accepted[string]{}

func (err Moved[V]) HTTPStatus() int {
	return http.StatusMovedPermanently
}
func (err Moved[V]) HTTPStatuses() []int {
	return []int{http.StatusMovedPermanently}
}
func (err Moved[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Result)
}
func (err *Moved[V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &err.Result)
}
func (err Moved[V]) APISchemaType() any {
	return err.Result
}
func (err Moved[V]) APIName() string {
	return getResultAPIName(err.Result, "Moved")
}

// A 400 response.
type BadRequest[V any] struct {
	Result V
}

func NewBadRequest[V any](result V) *BadRequest[V] {
	return &BadRequest[V]{result}
}

var _ invalidResult = &BadRequest[string]{}

func (err BadRequest[V]) HTTPStatus() int {
	return http.StatusBadRequest
}
func (err BadRequest[V]) HTTPStatuses() []int {
	return []int{http.StatusBadRequest}
}
func (err BadRequest[V]) Error() string {
	return getResultError(err.Result, err.HTTPStatus())
}
func (err BadRequest[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Result)
}
func (err *BadRequest[V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &err.Result)
}
func (err BadRequest[V]) APISchemaType() any {
	return err.Result
}
func (err BadRequest[V]) APIName() string {
	return getResultAPIName(err.Result, "BadRequest")
}

// A 401 response.
type Unauthorized[V any] struct {
	Result V
}

func NewUnauthorized[V any](result V) *Unauthorized[V] {
	return &Unauthorized[V]{result}
}

var _ invalidResult = &Unauthorized[string]{}

func (err Unauthorized[V]) HTTPStatus() int {
	return http.StatusUnauthorized
}
func (err Unauthorized[V]) HTTPStatuses() []int {
	return []int{http.StatusUnauthorized}
}
func (err Unauthorized[V]) Error() string {
	return getResultError(err.Result, err.HTTPStatus())
}
func (err Unauthorized[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Result)
}
func (err *Unauthorized[V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &err.Result)
}
func (err Unauthorized[V]) APISchemaType() any {
	return err.Result
}
func (err Unauthorized[V]) APIName() string {
	return getResultAPIName(err.Result, "Unauthorized")
}

// A 402 response.
type PaymentRequired[V any] struct {
	Result V
}

func NewPaymentRequired[V any](result V) *PaymentRequired[V] {
	return &PaymentRequired[V]{result}
}

var _ invalidResult = &PaymentRequired[string]{}

func (err PaymentRequired[V]) HTTPStatus() int {
	return http.StatusPaymentRequired
}
func (err PaymentRequired[V]) HTTPStatuses() []int {
	return []int{http.StatusPaymentRequired}
}
func (err PaymentRequired[V]) Error() string {
	return getResultError(err.Result, err.HTTPStatus())
}
func (err PaymentRequired[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Result)
}
func (err *PaymentRequired[V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &err.Result)
}
func (err PaymentRequired[V]) APISchemaType() any {
	return err.Result
}
func (err PaymentRequired[V]) APIName() string {
	return getResultAPIName(err.Result, "PaymentRequired")
}

// A 403 response.
type Forbidden[V any] struct {
	Result V
}

func NewForbidden[V any](result V) *Forbidden[V] {
	return &Forbidden[V]{result}
}

var _ invalidResult = &Forbidden[string]{}

func (err Forbidden[V]) HTTPStatus() int {
	return http.StatusForbidden
}
func (err Forbidden[V]) HTTPStatuses() []int {
	return []int{http.StatusForbidden}
}
func (err Forbidden[V]) Error() string {
	return getResultError(err.Result, err.HTTPStatus())
}
func (err Forbidden[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Result)
}
func (err *Forbidden[V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &err.Result)
}
func (err Forbidden[V]) APISchemaType() any {
	return err.Result
}
func (err Forbidden[V]) APIName() string {
	return getResultAPIName(err.Result, "Forbidden")
}

// A 404 response.
type NotFound[V any] struct {
	Result V
}

func NewNotFound[V any](result V) *NotFound[V] {
	return &NotFound[V]{result}
}

var _ invalidResult = &NotFound[string]{}

func (err NotFound[V]) HTTPStatus() int {
	return http.StatusNotFound
}
func (err NotFound[V]) HTTPStatuses() []int {
	return []int{http.StatusNotFound}
}
func (err NotFound[V]) Error() string {
	return getResultError(err.Result, err.HTTPStatus())
}
func (err NotFound[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Result)
}
func (err *NotFound[V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &err.Result)
}
func (err NotFound[V]) APISchemaType() any {
	return err.Result
}
func (err NotFound[V]) APIName() string {
	return getResultAPIName(err.Result, "NotFound")
}

// A 409 response.
type Conflict[V any] struct {
	Result V
}

func NewConflict[V any](result V) *Conflict[V] {
	return &Conflict[V]{result}
}

var _ invalidResult = &Conflict[string]{}

func (err Conflict[V]) HTTPStatus() int {
	return http.StatusConflict
}
func (err Conflict[V]) HTTPStatuses() []int {
	return []int{http.StatusConflict}
}
func (err Conflict[V]) Error() string {
	return getResultError(err.Result, err.HTTPStatus())
}
func (err Conflict[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Result)
}
func (err *Conflict[V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &err.Result)
}
func (err Conflict[V]) APISchemaType() any {
	return err.Result
}
func (err Conflict[V]) APIName() string {
	return getResultAPIName(err.Result, "Conflict")
}

// A 429 response.
type TooManyRequests[V any] struct {
	Result V
}

func NewTooManyRequests[V any](result V) *TooManyRequests[V] {
	return &TooManyRequests[V]{result}
}

var _ invalidResult = &TooManyRequests[string]{}

func (err TooManyRequests[V]) HTTPStatus() int {
	return http.StatusTooManyRequests
}
func (err TooManyRequests[V]) HTTPStatuses() []int {
	return []int{http.StatusTooManyRequests}
}
func (err TooManyRequests[V]) Error() string {
	return getResultError(err.Result, err.HTTPStatus())
}
func (err TooManyRequests[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Result)
}
func (err *TooManyRequests[V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &err.Result)
}
func (err TooManyRequests[V]) APISchemaType() any {
	return err.Result
}
func (err TooManyRequests[V]) APIName() string {
	return getResultAPIName(err.Result, "TooManyRequests")
}

// A 500 response.
type InternalServerError[V any] struct {
	Result V
}

func NewInternalServerError[V any](result V) *InternalServerError[V] {
	return &InternalServerError[V]{result}
}

var _ invalidResult = &InternalServerError[string]{}

func (err InternalServerError[V]) HTTPStatus() int {
	return http.StatusInternalServerError
}
func (err InternalServerError[V]) HTTPStatuses() []int {
	return []int{http.StatusInternalServerError}
}
func (err InternalServerError[V]) Error() string {
	return getResultError(err.Result, err.HTTPStatus())
}
func (err InternalServerError[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Result)
}
func (err *InternalServerError[V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &err.Result)
}
func (err InternalServerError[V]) APISchemaType() any {
	return err.Result
}
func (err InternalServerError[V]) APIName() string {
	return getResultAPIName(err.Result, "InternalServerError")
}

// A 501 response.
type NotImplemented[V any] struct {
	Result V
}

func NewNotImplemented[V any](result V) *NotImplemented[V] {
	return &NotImplemented[V]{result}
}

var _ invalidResult = &NotImplemented[string]{}

func (err NotImplemented[V]) HTTPStatus() int {
	return http.StatusNotImplemented
}
func (err NotImplemented[V]) HTTPStatuses() []int {
	return []int{http.StatusNotImplemented}
}
func (err NotImplemented[V]) Error() string {
	return getResultError(err.Result, err.HTTPStatus())
}
func (err NotImplemented[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Result)
}
func (err *NotImplemented[V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &err.Result)
}
func (err NotImplemented[V]) APISchemaType() any {
	return err.Result
}
func (err NotImplemented[V]) APIName() string {
	return getResultAPIName(err.Result, "NotImplemented")
}

// A 503 response.
type ServiceUnavailable[V any] struct {
	Result V
}

func NewServiceUnavailable[V any](result V) *ServiceUnavailable[V] {
	return &ServiceUnavailable[V]{result}
}

var _ invalidResult = &ServiceUnavailable[string]{}

func (err ServiceUnavailable[V]) HTTPStatus() int {
	return http.StatusServiceUnavailable
}
func (err ServiceUnavailable[V]) HTTPStatuses() []int {
	return []int{http.StatusServiceUnavailable}
}
func (err ServiceUnavailable[V]) Error() string {
	return getResultError(err.Result, err.HTTPStatus())
}
func (err ServiceUnavailable[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(err.Result)
}
func (err *ServiceUnavailable[V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &err.Result)
}
func (err ServiceUnavailable[V]) APISchemaType() any {
	return err.Result
}
func (err ServiceUnavailable[V]) APIName() string {
	return getResultAPIName(err.Result, "ServiceUnavailable")
}

// A custom status response. The documentation will not be able to
// report on the status of this type, you need to use one of the provided
// types or define your own implementor of HasStatus.
type Result[V any] struct {
	Status int
	Value  V
}

func NewResult[V any](status int, result V) *Result[V] {
	return &Result[V]{status, result}
}

var _ json.Marshaler = &Result[string]{}
var _ json.Unmarshaler = &Result[string]{}
var _ HasStatus = &Result[string]{}
var _ api.HasSchemaType = Result[string]{}

func (se Result[V]) HTTPStatus() int {
	return se.Status
}
func (se Result[V]) HTTPStatuses() []int {
	return []int{}
}
func (se Result[V]) MarshalJSON() ([]byte, error) {
	return json.Marshal(se.Value)
}
func (se *Result[V]) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &se.Value)
}
func (err Result[V]) APISchemaType() any {
	return err.Value
}
func (err Result[V]) APIName() string {
	return api.GetName(err.Value)
}

// Converts the result to an error string or if it has no string
// representation it returns the text of the status provided.
func getResultError(result any, status int) string {
	msg := fmt.Sprintf("%v", result)
	if msg == "" {
		msg = http.StatusText(status)
	}
	return msg
}

// Returns the APIName for the given type. If it implements HasName
// that will be used as the type name. Otherwise prefix will be added
// to the beginning of the Type name.
func getResultAPIName(result any, prefix string) string {
	valueType := reflect.TypeOf(result)
	name := api.GetName(valueType)
	if api.IsNamedType(valueType) {
		return name
	}
	return prefix + name
}
