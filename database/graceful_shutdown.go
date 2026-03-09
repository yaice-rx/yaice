package database

import (
	"os"
	"os/signal"
	"syscall"
)

func WaitForShutdown(shutdown func()) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)

	<-ch
	shutdown()
}
