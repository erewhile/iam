package cmd

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/erewhile/iam/cmd/flags"
	"github.com/erewhile/iam/config"
	"github.com/erewhile/iam/internal/router"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"golang.org/x/net/netutil"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "iam server",
	Run: func(cmd *cobra.Command, args []string) {
		server()
	},
}

func server() {
	if err := writePID(); err != nil {
		log.Fatalf("failed to write PID file: %v\n", err)
	}
	defer removePID()

	setup()
	defer release()

	if !flags.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()
	router.Init(r)

	addr := config.Get().Scheme.Port
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on address %s: %s\n", addr, err)
	}
	limitedLn := netutil.LimitListener(ln, 1<<11)

	srv := &http.Server{
		Handler:           r,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       30 * time.Second,
		MaxHeaderBytes:    1 << 20, // 1MB
	}

	go func() {
		if err := srv.Serve(limitedLn); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v\n", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt,    // Interrupt signal (CTRL+C)
		syscall.SIGINT,  // SIGINT signal
		syscall.SIGTERM, // SIGTERM signal (Termination)
		syscall.SIGQUIT, // SIGQUIT signal (Quit)
	)
	defer stop()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown failed, forcing exit: %v\n", err)
	}
}

func init() {
	rootCmd.AddCommand(serverCmd)
}
