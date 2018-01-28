package reply

import (
	"encoding/json"
	"net/http"

	"github.com/leopoldxx/go-utils/errors"
)

type commResp struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// CommReply can be used for replying some common data
func CommReply(w http.ResponseWriter, r *http.Request, status int, message string) {
	resp := commResp{
		Code:    http.StatusText(status),
		Message: message,
	}
	Reply(w, r, status, resp)
}

// Reply can be used for replying response
// Rename this to "reply" in the future because should call ResponseReply instead
func Reply(w http.ResponseWriter, r *http.Request, status int, v interface{}) {
	data, _ := json.Marshal(v)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}

// OK reply
func OK(w http.ResponseWriter, r *http.Request, message string) {
	CommReply(w, r, http.StatusOK, message)
}

// ResourceNotFound will return an error message indicating that the resource is not exist
func ResourceNotFound(w http.ResponseWriter, r *http.Request, message string) {
	CommReply(w, r, http.StatusNotFound, message)
}

// BadRequest will return an error message indicating that the request is invalid
func BadRequest(w http.ResponseWriter, r *http.Request, err error) {
	CommReply(w, r, http.StatusBadRequest, err.Error())
}

// Forbidden will block user access the resource, not authorized
func Forbidden(w http.ResponseWriter, r *http.Request, err error) {
	CommReply(w, r, http.StatusForbidden, err.Error())
}

// Unauthorized will block user access the api, not login
func Unauthorized(w http.ResponseWriter, r *http.Request, err error) {
	CommReply(w, r, http.StatusUnauthorized, err.Error())
}

// InternalError will return an error message indicating that the something is error inside the controller
func InternalError(w http.ResponseWriter, r *http.Request, err error) {
	CommReply(w, r, http.StatusInternalServerError, err.Error())
}

// ServiceUnavailable will return an error message indicating that the service is not available now
func ServiceUnavailable(w http.ResponseWriter, r *http.Request, err error) {
	CommReply(w, r, http.StatusServiceUnavailable, err.Error())
}

// Conflict xxx
func Conflict(w http.ResponseWriter, r *http.Request, err error) {
	CommReply(w, r, http.StatusConflict, err.Error())
}

// NotAcceptable xxx
func NotAcceptable(w http.ResponseWriter, r *http.Request, err error) {
	CommReply(w, r, http.StatusNotAcceptable, err.Error())
}

// SetRequestID will set the response header of the requestID
func SetRequestID(w http.ResponseWriter, requestID string) {
	w.Header().Set("x-request-id", requestID)
}

// ProcessError xxx
func ProcessError(w http.ResponseWriter, r *http.Request, err error) {
	if errors.IsNotFoundError(err) {
		ResourceNotFound(w, r, err.Error())
	} else if errors.IsConflictError(err) {
		Conflict(w, r, err)
	} else if errors.IsNotReadyError(err) || errors.IsTaskIsRunningError(err) {
		NotAcceptable(w, r, err)
	} else if errors.IsBadRequestError(err) || errors.IsClientError(err) || errors.IsInvalidRegionError(err) {
		BadRequest(w, r, err)
	} else if errors.IsForbiddenError(err) {
		Forbidden(w, r, err)
	} else {
		InternalError(w, r, err)
	}
}
