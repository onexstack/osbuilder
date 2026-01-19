package errno

import (
	"net/http"

	"github.com/onexstack/onexstack/pkg/errorsx"
)

// ErrInvalidMessageType indicates that the specified message type was invalid.
var ErrInvalidMessageType = errorsx.New(http.StatusBadRequest, "InvalidArgument.InvalidMessageType", "Message type is invalid.")

// ErrPayloadInvalid indicates that the specified message payload was invalid.
var ErrPayloadInvalid = errorsx.New(http.StatusBadRequest, "InvalidArgument.PayloadInvalid", "Message payload invalid.")

// ErrPing indicates that the ping message is failed.
var ErrPing = errorsx.New(http.StatusInternalServerError, "InternalError.PingFailed", "Ping message failed.")
