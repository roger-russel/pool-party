package pool

import "errors"

var errAddONShuttingDown = errors.New("pool is shuttingdown no new tasks will be accept")
