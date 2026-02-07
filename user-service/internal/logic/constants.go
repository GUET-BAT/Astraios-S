package logic

import "time"

// Response codes used in user service RPC responses.
const (
	CodeSuccess      int32 = 0
	CodeInvalidParam int32 = 1
	CodeAlreadyExists int32 = 2
	CodeInternal     int32 = 3
)

// Database query timeout for all SQL operations.
const dbQueryTimeout = 5 * time.Second
