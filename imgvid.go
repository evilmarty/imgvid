package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

var (
	dateHeaders = []string{"last-modified", "date"}
)

type Imgvid struct {
	URL        *url.URL
	Response   *http.Response
	WorkingDir string
	Rate       int
	MaxFrames  int
}

func NewImgvid(u *url.URL) *Imgvid {
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(u.String())))
	workingDir := filepath.Join(cachePath, hash)

	return &Imgvid{
		URL:        u,
		WorkingDir: workingDir,
		Rate:       defaultRate,
		MaxFrames:  defaultMaxFrames,
	}
}

func NewImgvidFromValues(values url.Values) (*Imgvid, error) {
	if u, err := getUrlParam(values, "url"); err != nil {
		return nil, err
	} else {
		imgvid := NewImgvid(u)
		if rate, err := getIntParam(values, "rate"); rate > 0 && err == nil {
			imgvid.Rate = rate
		}
		if max, err := getIntParam(values, "max"); err == nil {
			imgvid.MaxFrames = max
		}
		return imgvid, nil
	}
}

func (imgvid *Imgvid) ContentType() string {
	return imgvid.Response.Header.Get("content-type")
}

func (imgvid *Imgvid) IsCompatible() bool {
	contentType := imgvid.ContentType()

	for _, mt := range compatibleMimeTypes {
		if mt == contentType {
			return true
		}
	}
	return false
}

func (imgvid *Imgvid) Extension() string {
	if exts, err := mime.ExtensionsByType(imgvid.ContentType()); err != nil {
		return ""
	} else {
		return exts[0]
	}
}

func (imgvid *Imgvid) Codec() string {
	ext := imgvid.Extension()
	return ext[1:]
}

func (imgvid *Imgvid) LastModified() time.Time {
	for _, header := range dateHeaders {
		if t, e := http.ParseTime(imgvid.Response.Header.Get(header)); e == nil {
			return t
		}
	}

	return time.Now()
}

func (imgvid *Imgvid) Request() error {
	if resp, err := http.Get(imgvid.URL.String()); err != nil {
		return err
	} else {
		imgvid.Response = resp
	}

	if !imgvid.IsCompatible() {
		return fmt.Errorf("Unsupported mimetype: %s", imgvid.ContentType())
	}

	return nil
}

func (imgvid *Imgvid) Write(src io.Reader) (int64, error) {
	if err := os.MkdirAll(imgvid.WorkingDir, 0744); err != nil {
		return 0, err
	}

	lastModified := imgvid.LastModified()
	ext := imgvid.Extension()
	path := filepath.Join(imgvid.WorkingDir, fmt.Sprintf("%d%s", lastModified.Unix(), ext))

	if file, fileErr := os.Create(path); fileErr != nil {
		return 0, fileErr
	} else {
		defer file.Close()
		return io.Copy(file, src)
	}
}

func (imgvid *Imgvid) Get() error {
	var err error

	if err = imgvid.Request(); err != nil {
		return err
	}

	defer imgvid.Response.Body.Close()

	if _, err = imgvid.Write(imgvid.Response.Body); err != nil {
		return err
	}

	return nil
}

func (imgvid *Imgvid) GlobPattern() string {
	ext := imgvid.Extension()
	return filepath.Join(imgvid.WorkingDir, fmt.Sprintf("*%s", ext))
}

func (imgvid *Imgvid) Download(w http.ResponseWriter) error {
	args := append(ffmpegExtraArgs, "-r", strconv.Itoa(imgvid.Rate), "-codec", imgvid.Codec(), "-i", imgvid.GlobPattern(), "-f", "mpeg", "pipe:1")
	cmd := exec.Command(ffmpegBinPath, args...)
	stdout, _ := cmd.StdoutPipe()

	w.Header().Set("content-type", "video/mpeg")

	if verbose {
		log.Printf("Invoking: %s", cmd.String())
		cmd.Stderr = log.Writer()
	}
	if err := cmd.Start(); err != nil {
		return err
	}

	go io.Copy(w, stdout)

	return cmd.Wait()
}

func (imgvid *Imgvid) Cleanup() error {
	if files, err := filepath.Glob(imgvid.GlobPattern()); err != nil {
		return err
	} else if len(files) > imgvid.MaxFrames {
		d := len(files) - imgvid.MaxFrames
		for _, file := range files[0:d] {
			if err := os.Remove(file); err != nil {
				return err
			}
		}
	}

	return nil
}
