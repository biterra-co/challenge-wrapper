package main

import (
	"bytes"
	_ "embed"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

//go:embed browser.js
var browserJS []byte

const scriptTag = `<script type="module" src="/__biterra__/challenge-wrapper.js"></script>`

func main() {
	if len(os.Args) < 2 {
		log.Fatal("challenge command is required")
	}
	port := getenv("BITERRA_UPSTREAM_PORT", "3000")
	listen := getenv("BITERRA_LISTEN_ADDR", ":8080")
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = os.Stdout, os.Stderr, os.Stdin
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}

	target, _ := url.Parse("http://127.0.0.1:" + port)
	proxy := httputil.NewSingleHostReverseProxy(target)
	baseDirector := proxy.Director
	proxy.Director = func(r *http.Request) { baseDirector(r); r.Header.Set("Accept-Encoding", "identity") }
	proxy.ModifyResponse = inject
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/__biterra__/challenge-wrapper.js" {
			w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
			w.Header().Set("Cache-Control", "public, max-age=300")
			_, _ = w.Write(browserJS)
			return
		}
		proxy.ServeHTTP(w, r)
	})
	server := &http.Server{Addr: listen, Handler: h}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("proxy: %v", err)
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() { sig := <-sigs; _ = syscall.Kill(-cmd.Process.Pid, sig.(syscall.Signal)) }()
	err := cmd.Wait()
	_ = server.Close()
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			os.Exit(exit.ExitCode())
		}
		log.Fatal(err)
	}
}

func inject(res *http.Response) error {
	if !strings.Contains(strings.ToLower(res.Header.Get("Content-Type")), "text/html") || res.Body == nil {
		return nil
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	_ = res.Body.Close()
	if bytes.Contains(body, []byte("/__biterra__/challenge-wrapper.js")) {
		res.Body = io.NopCloser(bytes.NewReader(body))
		return nil
	}
	needle := []byte("</head>")
	i := bytes.Index(bytes.ToLower(body), needle)
	if i < 0 {
		needle = []byte("</body>")
		i = bytes.Index(bytes.ToLower(body), needle)
	}
	if i < 0 {
		body = append(body, []byte(scriptTag)...)
	} else {
		body = append(body[:i], append([]byte(scriptTag), body[i:]...)...)
	}
	res.Body = io.NopCloser(bytes.NewReader(body))
	res.ContentLength = -1
	res.Header.Del("Content-Length")
	res.Header.Del("Content-Encoding")
	return nil
}

func getenv(k, fallback string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return fallback
}
