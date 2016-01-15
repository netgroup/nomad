package driver

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	bufSize = 32 * 1024
)

type LogRotator struct {
	maxFiles   int
	fileSize   int64
	path       string
	fileName   string
	logFileIdx int

	logger *log.Logger
}

func NewLogRotator(path string, fileName string, maxFiles int, fileSize int64, logger *log.Logger) (*LogRotator, error) {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	logFileIdx := 0
	for _, f := range files {
		if strings.HasPrefix(f.Name(), fileName) {
			fileIdx := strings.TrimPrefix(f.Name(), fmt.Sprintf("%s.", fileName))
			n, err := strconv.Atoi(fileIdx)
			if err != nil {
				continue
			}
			if n > logFileIdx {
				logFileIdx = n
			}
		}
	}

	return &LogRotator{
		maxFiles:   maxFiles,
		fileSize:   fileSize,
		path:       path,
		fileName:   fileName,
		logFileIdx: logFileIdx,
		logger:     logger,
	}, nil
}

func (l *LogRotator) Start(r io.Reader) error {
	buf := make([]byte, bufSize)
	for {
		logFileName := filepath.Join(l.path, fmt.Sprintf("%s.%d", l.fileName, l.logFileIdx))
		if f, err := os.Stat(logFileName); err == nil {
			if f.IsDir() {
				l.logFileIdx += 1
				continue
			}
		}
		f, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return err
		}
		l.logger.Println("[INFO] logrotator: opened a new file: %s", logFileName)
		remainingSize := l.fileSize
		if finfo, err := os.Stat(logFileName); err == nil {
			remainingSize -= finfo.Size()
		}
		if remainingSize < 1 {
			l.logFileIdx = l.logFileIdx + 1
			f.Close()
			continue
		}

		for {
			var nr int
			var err error
			if remainingSize < bufSize {
				nr, err = r.Read(buf[0:remainingSize])
			} else {
				nr, err = r.Read(buf)
			}
			if err != nil {
				f.Close()
				return err
			}
			nw, err := f.Write(buf[:nr])
			if err != nil {
				f.Close()
				return err
			}
			if nr != nw {
				f.Close()
				return fmt.Errorf("Failed to write data R: %d W: %d", nr, nw)
			}
			remainingSize -= int64(nr)
			if remainingSize < 1 {
				f.Close()
				break
			}
		}
		l.logFileIdx = l.logFileIdx + 1
	}
	return nil
}