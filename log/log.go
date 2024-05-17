package log

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	DebugMode   = false
	logFilePath string
	LogFile     *os.File
)

type Requester struct {
	Address string
	Agent   string
	Url     string
}

const (
	WebSocketOpen  = 1
	WebSocketClose = 2
)

func (r *Requester) Make(request *http.Request) {
	r.Address = request.RemoteAddr
	r.Agent = request.UserAgent()
	r.Url = request.URL.Path
}

func SetLogFile(logFile *os.File) {
	logFilePath = logFile.Name()
	LogFile = logFile
	Debug("logging to file: %s", logFilePath)
	log.SetOutput(logFile)
}

func BackupLogFile(date string) {
	if logFilePath == "" {
		return
	}
	newFile := logFilePath + "." + date + ".gz"
	newFileObj, err := os.Create(newFile)
	if err != nil {
		Error(err)
		return
	}
	defer newFileObj.Close()

	writer := gzip.NewWriter(newFileObj)
	defer writer.Close()

	oldFile, err := os.Open(logFilePath)
	if err != nil {
		Error(err)
		return
	}
	defer oldFile.Close()

	_, err = io.Copy(writer, oldFile)
	if err != nil {
		Error(err)
		return
	}

	err = os.Remove(logFilePath)
	if err != nil {
		Error(err)
		return
	}

	newLogFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		Error(err)
		return
	}
	log.SetOutput(newLogFile)
	Info("backed up log file to %s", newFile)

	// delete old log files that are older than 7 days
	files, err := filepath.Glob(logFilePath + ".*.gz")
	if err != nil {
		Error(err)
		return
	}
	for _, file := range files {
		fileInfo, err := os.Stat(file)
		if err != nil {
			Error(err)
			continue
		}
		if time.Since(fileInfo.ModTime()) > 7*24*time.Hour {
			err = os.Remove(file)
			if err != nil {
				Error(err)
			}
		}
	}
}

func _log(t string, caller int, msg string) {
	// get info about the caller
	pc, _, line, _ := runtime.Caller(caller)
	fn := runtime.FuncForPC(pc).Name()

	// split the file path and get the function name
	// (they will be separated by a period)
	// i.e. /path/to/file.functionName
	paths := strings.Split(fn, ".")
	l := len(paths)

	// get the function name and the file name
	functionName := paths[l-1]
	// cannot use 0 because the first element might have a period
	// in the path, so use length - 2
	fileName := filepath.Base(paths[l-2])
	log.Printf("[%s] %s.%s:%d %s", t, fileName, functionName, line, msg)
}

// log the error with the file and line number
func Error(err error) error {
	if err != nil {
		_log("error", 2, err.Error())
	}
	return err
}

func Errorf(format string, args ...interface{}) error {
	err := fmt.Errorf(format, args...)
	_log("error", 2, err.Error())
	return err
}

// log the message with the file and line number
func Info(format string, args ...interface{}) {
	_log("info", 2, fmt.Sprintf(format, args...))
}

func Debug(format string, args ...interface{}) {
	if DebugMode {
		_log("debug", 2, fmt.Sprintf(format, args...))
	}
}

func _warn(format string, args ...interface{}) {
	_log("warn", 3, fmt.Sprintf(format, args...))
}

func Warn(format string, args ...interface{}) {
	_warn(format, args...)
}

func Warning(format string, args ...interface{}) {
	_warn(format, args...)
}

// log the http request with the status code
func Http(status int, r Requester) {
	log.Printf("[http] (%d) %s from %s (%s)", status, r.Url, r.Address, r.Agent)
}

func WS(status int, r Requester) {
	if status == WebSocketOpen {
		log.Printf("[webs] (%s/%s) session opened", r.Address, r.Agent)
	} else if status == WebSocketClose {
		log.Printf("[webs] (%s/%s) session closed", r.Address, r.Agent)
	}
}

func WSMessage(r Requester, msg string) {
	log.Printf("[webs] (%s/%s) %s", r.Address, r.Agent, msg)
}
