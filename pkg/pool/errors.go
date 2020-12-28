package pool

import "errors"

var errAddONShuttingDown = errors.New("pool is shuttingdown no new tasks will be accept")
var errQueuedTasksWillBeExecutedNoMore = errors.New("queued tasks will no longer be performed due to shutdown")
