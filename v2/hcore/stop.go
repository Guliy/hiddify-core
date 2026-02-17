package hcore

import (
	"context"
	"fmt"
	"time"

	"github.com/hiddify/hiddify-core/v2/config"
	hcommon "github.com/hiddify/hiddify-core/v2/hcommon"
)

func (s *CoreService) Stop(ctx context.Context, empty *hcommon.Empty) (*CoreInfoResponse, error) {
	return Stop()
}

func Stop() (coreResponse *CoreInfoResponse, err error) {
	defer config.DeferPanicToError("stop", func(recovered_err error) {
		coreResponse, err = errorWrapper(MessageType_UNEXPECTED_ERROR, recovered_err)
	})

	static.lock.Lock()
	defer static.lock.Unlock()

	SetCoreStatus(CoreStates_STOPPING, MessageType_EMPTY, "")
	ss := static.StartedService
	if ss == nil {
		cleanupStaleIPRules()
		return SetCoreStatus(CoreStates_STOPPED, MessageType_ALREADY_STOPPED, ""), nil
	}

	// CloseService can hang if TUN/nftables teardown blocks — use a timeout
	done := make(chan error, 1)
	go func() {
		done <- ss.CloseService()
	}()

	select {
	case closeErr := <-done:
		if closeErr != nil {
			Log(LogLevel_WARNING, LogType_CORE, "CloseService error: ", closeErr.Error())
			dumpGoroutinesToFile(fmt.Sprint(sWorkingPath, "/data/goroutine-stop.log"))
		}
	case <-time.After(5 * time.Second):
		Log(LogLevel_WARNING, LogType_CORE, "CloseService timed out after 5s, forcing cleanup")
		dumpGoroutinesToFile(fmt.Sprint(sWorkingPath, "/data/goroutine-stop.log"))
	}

	static.StartedService = nil
	// Always clean up ip rules and nftables after stop
	cleanupStaleIPRules()

	return SetCoreStatus(CoreStates_STOPPED, MessageType_EMPTY, ""), nil
}
