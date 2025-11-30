package errno

import (
	"net/http"

	"github.com/onexstack/onexstack/pkg/errorsx"
)

// ErrInvalidMessageType indicates that the specified message type was invalid.
var ErrInvalidMessageType = &errorsx.ErrorX{
	Code:    http.StatusBadRequest,
	Reason:  "InvalidArgument.InvalidMessageType",
	Message: "Message type is invalid.",
}

// ErrPayloadInvalid indicates that the specified message payload was invalid.
var ErrPayloadInvalid = &errorsx.ErrorX{
	Code:    http.StatusBadRequest,
	Reason:  "InvalidArgument.PayloadInvalid",
	Message: "Message payload invalid.",
}

// ErrPing indicates that the ping message is failed.
var ErrPing = &errorsx.ErrorX{
	Code:    http.StatusInternalServerError,
	Reason:  "InternalError.PingFailed",
	Message: "Ping message failed.",
}
