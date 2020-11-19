package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	appName = "imgvid"
)

var (
	compatibleMimeTypes = []string{"image/jpg", "image/gif", "image/png"}
	ffmpegExtraArgs     = []string{"-loglevel", "info", "-f", "image2", "-pattern_type", "glob"}
	ffmpegBinPath       = "ffmpeg"
	cachePath           = ""
	port                = 8080
	verbose             = false
	defaultRate         = 5
	defaultMaxFrames    = 10
)

func init() {
	var err error

	if cachePath, err = os.UserCacheDir(); err != nil {
		panic(err)
	} else {
		cachePath = filepath.Join(cachePath, appName)
	}

	flag.StringVar(&ffmpegBinPath, "ffmpeg", ffmpegBinPath, "Path to ffmpeg command.")
	flag.StringVar(&cachePath, "cache", cachePath, "Path to store cache files. Defaults to user cache directory.")
	flag.IntVar(&port, "port", port, "Port used by the http server.")
	flag.IntVar(&defaultRate, "rate", defaultRate, "Default video framerate.")
	flag.IntVar(&defaultMaxFrames, "max", defaultMaxFrames, "Default maximum frames.")
	flag.BoolVar(&verbose, "verbose", verbose, "Verbose output.")
	flag.Parse()

	if ffmpegBinPath, err = exec.LookPath(ffmpegBinPath); err != nil {
		panic(err)
	}
}

func main() {
	addr := fmt.Sprintf(":%d", port)
	log.Printf("Using cache directory: %s\n", cachePath)
	log.Printf("Server listening on %s...", addr)

	http.HandleFunc("/_health", healthcheck)
	http.HandleFunc("/", process)
	http.ListenAndServe(addr, nil)
}

func process(w http.ResponseWriter, req *http.Request) {
	log.Printf("%s \"%s %s\" - %s", req.RemoteAddr, req.Method, req.RequestURI, req.UserAgent())

	if req.Method != "GET" {
		http.Error(w, "Method Not Allowed", 405)
		return
	} else if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}

	if imgvid, err := NewImgvidFromValues(req.URL.Query()); err != nil {
		http.Error(w, err.Error(), 400)
	} else if err := imgvid.Get(); err != nil {
		http.Error(w, err.Error(), 400)
	} else if err := imgvid.Download(w); err != nil {
		http.Error(w, err.Error(), 500)
	} else if err := imgvid.Cleanup(); err != nil && verbose {
		log.Printf("Clean up error: %v", err)
	}
}

func healthcheck(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintf(w, "ok")
}
