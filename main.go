package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Site struct {
	Name            string `yaml:"name"`
	Prefix          string `yaml:"prefix"`
	Offset          int    `yaml:"offset"`
	DhcpServerType  string `yaml:"dhcp_server_type"`
	DhcpServer      string `yaml:"dhcp_server"`
	DhcpApiPort     int    `yaml:"dhcp_api_port"`
	Community       string `yaml:"community"`
	DhcpApiLogin    string `yaml:"login"`
	DhcpApiPassword string `yaml:"password"`
}

func loggingMiddleware(next http.Handler) http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, next)
}

func run(ctx context.Context) (err error) {
	var (
		cfg *Config
	)
	if cfg, err = NewConfig("config.yaml"); err != nil {
		log.Fatal(err)
	} else {
		router := mux.NewRouter().StrictSlash(true)
		router.Use(loggingMiddleware)
		addrString := cfg.Service.Listen + ":" + cfg.Service.Port
		//c := NbNewClient(cfg.Netbox.Address, cfg.Netbox.Token)

		srv := &http.Server{
			Addr:    addrString,
			Handler: router,
		}
		go func() {
			if err = srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen:%+s\n", err)
			}
		}()
		log.Printf("server started")
		<-ctx.Done()

		log.Printf("server stopped")
		ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer func() {
			cancel()
		}()

		if err = srv.Shutdown(ctxShutDown); err != nil {
			log.Fatalf("server Shutdown Failed:%+s", err)
		}

		log.Printf("server exited properly")

		if err == http.ErrServerClosed {
			err = nil
		}

	}
	return
}

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		oscall := <-c
		log.Printf("system call:%+v", oscall)
		cancel()
	}()

	if err := run(ctx); err != nil {
		log.Printf("failed to serve:+%v\n", err)
	}
}
