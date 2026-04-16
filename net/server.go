package net

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/golang-devkit/pkg/logger"
	"go.uber.org/zap"
)

func GetWd() string {
	wdPath, err := os.Getwd()
	if err != nil {
		log.Printf("can not get path of current directory: %s", err.Error())
		return "."
	}
	return wdPath
}

func FileServer(prefix, dirPath string) http.Handler {
	fs := http.FileServer(http.Dir(dirPath))
	return http.StripPrefix(prefix, fs)
}

func HttpServerWithConfig(addr string, handler http.Handler) *http.Server {
	//
	def := &http.Server{
		Addr:         "0.0.0.0:8080",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}
	//
	defer func() {
		logger.NewEntry().Debug("Starting HTTP server",
			zap.String("working_directory", fmt.Sprintf("%s%s", GetWd(), string(filepath.Separator))),
			zap.String("address", def.Addr),
		)
	}()
	//
	if addr != "" && addr != def.Addr {
		if strings.HasPrefix(addr, ":") {
			def.Addr = "0.0.0.0" + addr
		} else {
			def.Addr = addr
		}
	}
	return def
}
