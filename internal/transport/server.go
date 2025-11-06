package transport

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server interface {
	Run() error
	Shutdown(context.Context) error
}

func RunServers(servers ...Server) error {
	errCh := make(chan error)

	for _, server := range servers {
		go func(s Server) {
			if err := s.Run(); err != nil {
				err = fmt.Errorf("error starting REST server: %w", err)
				errCh <- err
			}
		}(server)
	}

	notifyCh := make(chan os.Signal, 1)
	signal.Notify(notifyCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Println("Shutting down due to server error:", err)
		return shutdownServer(servers...)
	case <-notifyCh:
		log.Println("Shutting down gracefully...")
		_ = shutdownServer(servers...)
		return nil
	}
}

func shutdownServer(servers ...Server) (err error) {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for _, server := range servers {
		if se := server.Shutdown(shutdownCtx); se != nil {
			log.Println("Error during server shutdown:", se)
			err = errors.Join(err, se)
		}
	}

	if err != nil {
		return err
	}

	return nil
}
