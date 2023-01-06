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
)

func main() {
	upstream := "localhost:5000"
	listenAddr := ":8000"
	fmt.Println("starting adapter: ", listenAddr, " -> ", upstream)
	mux := http.NewServeMux()
	s := http.Server{
		Addr:    listenAddr,
		Handler: mux,
	}
	mux.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.DefaultClient.Get(upstream)
		if err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode > 299 {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		w.Header().Add("content-type", "application/json")
		w.Write([]byte(`{"state": "healthy", "gpu": true}`))
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		resp, err := http.DefaultClient.Post(upstream+"/predictions", "application/json", r.Body)
		if err != nil {
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
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
	go s.ListenAndServe()
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		fmt.Println("got interrupt signal")
		s.Shutdown(context.Background())
		os.Exit(0)
	}()
	cmd.Run()
	fmt.Println("command finished - shutting down adapter")
	s.Shutdown(context.Background())
}
