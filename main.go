package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	upstream := "http://localhost:5000"
	listenAddr := ":8000"
	fmt.Println("starting adapter: ", listenAddr, " -> ", upstream)
	mux := http.NewServeMux()
	s := http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}
	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.DefaultClient.Get(upstream + "/")
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode > 299 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.Header().Add("content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"state": "healthy", "gpu": true}`))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		resp, err := http.DefaultClient.Post(upstream+"/predictions", "application/json", r.Body)
		if err != nil {
			fmt.Println("couldn't reach upstream", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		for k, vs := range resp.Header {
			r.Header.Set(k, vs[0])
		}
		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, resp.Body)
		if err != nil {
			fmt.Println("couldn't read upstream", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	go func() {
		// we must not start listening before upstream is ready as
		// the health-check function above seems to get ignored
		for {
			_, err := http.DefaultClient.Get(upstream + "/")
			if err == nil {
				break
			}
			time.Sleep(100 * time.Millisecond)
		}
		s.ListenAndServe()
	}()
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		fmt.Println("got interrupt signal")
		s.Shutdown(context.Background())
		os.Exit(0)
	}()
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
	fmt.Println("command finished - shutting down adapter")
	s.Shutdown(context.Background())
}
