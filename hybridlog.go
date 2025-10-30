package tt

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

type HybridLogger struct {
	*logrus.Logger
	mu          sync.Mutex
	lumber      *lumberjack.Logger
	logDir      string
	fileName    string
	currentDate string
	timeFormat  string
}

// Level mapping for int → logrus.Level
var levelMap = map[int]logrus.Level{
	0: logrus.PanicLevel,
	1: logrus.FatalLevel,
	2: logrus.ErrorLevel,
	3: logrus.WarnLevel,
	4: logrus.InfoLevel,
	5: logrus.DebugLevel,
	6: logrus.TraceLevel,
}

// Init initializes the logger
// logDir: log directory
// logFileName: log file name
// maxSizeMB: max size of log file in MB, if exceeds, then it will rotate to new one
// maxBackups: max number of log files to keep, if exceeds, then it will delete the oldest log file
// maxAgeDays: max age of log files in days, if exceeds, then it will delete the oldest log file
// level: log level uint, 6:Trace, 5:Debug, 4:Info, 3:Warn, 2:Error, 1:Fatal, 0:Panic
// compress: whether to compress log files
func Init(logDir, logFileName string, maxSizeMB, maxBackups, maxAgeDays int, logLevel int, compress bool) (logObj *HybridLogger, err error) {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		err = fmt.Errorf("failed to create log dir: %v", err)
		return nil, err
	}
	timeFormat := "2006-01-02"
	// Get current date in YYYY-MM-DD format
	currentDate := time.Now().Format(timeFormat)
	ext := filepath.Ext(logFileName)
	nameWithoutExt := logFileName[:len(logFileName)-len(ext)]

	lumber := &lumberjack.Logger{
		Filename:   filepath.Join(logDir, fmt.Sprintf("%s-%s%s", nameWithoutExt, currentDate, ext)),
		MaxSize:    maxSizeMB,
		MaxBackups: maxBackups,
		MaxAge:     maxAgeDays,
		Compress:   compress,
	}

	h := &HybridLogger{
		Logger:      logrus.New(),
		lumber:      lumber,
		logDir:      logDir,
		fileName:    logFileName,
		currentDate: currentDate,
		timeFormat:  timeFormat,
	}

	h.Logger.SetOutput(h)
	h.SetLogLevel(logLevel) // Set initial level
	h.Logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	return h, nil
}

// Write sends logs to lumberjack for rotation
func (h *HybridLogger) Write(p []byte) (n int, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if date has changed
	currentDate := time.Now().Format(h.timeFormat)
	if h.currentDate != currentDate {
		// Close the current log file
		if h.lumber != nil {
			h.lumber.Close()
		}

		// Create a new log file with updated date
		ext := filepath.Ext(h.fileName)
		nameWithoutExt := h.fileName[:len(h.fileName)-len(ext)]
		h.lumber = &lumberjack.Logger{
			Filename:   filepath.Join(h.logDir, fmt.Sprintf("%s-%s%s", nameWithoutExt, currentDate, ext)),
			MaxSize:    h.lumber.MaxSize,
			MaxBackups: h.lumber.MaxBackups,
			MaxAge:     h.lumber.MaxAge,
			Compress:   h.lumber.Compress,
		}
		h.currentDate = currentDate
	}

	return h.lumber.Write(p)
}

// SetLogLevel changes log level at runtime (using int)
func (h *HybridLogger) SetLogLevel(level int) {
	if lvl, ok := levelMap[level]; ok {
		h.Logger.SetLevel(lvl)
	} else {
		h.Logger.SetLevel(logrus.InfoLevel) // default
	}
}

// --------- Wrapper Functions ---------

func (h *HybridLogger) Info(args ...interface{}) { h.Logger.Info(args...) }
func (h *HybridLogger) Infof(format string, args ...interface{}) {
	h.Logger.Infof(format, args...)
}

func (h *HybridLogger) Debug(args ...interface{}) { h.Logger.Debug(args...) }
func (h *HybridLogger) Debugf(format string, args ...interface{}) {
	h.Logger.Debugf(format, args...)
}

func (h *HybridLogger) Warn(args ...interface{}) { h.Logger.Warn(args...) }
func (h *HybridLogger) Warnf(format string, args ...interface{}) {
	h.Logger.Warnf(format, args...)
}

func (h *HybridLogger) Error(args ...interface{}) { h.Logger.Error(args...) }
func (h *HybridLogger) Errorf(format string, args ...interface{}) {
	h.Logger.Errorf(format, args...)
}

func (h *HybridLogger) Fatal(args ...interface{}) { h.Logger.Fatal(args...) }
func (h *HybridLogger) Fatalf(format string, args ...interface{}) {
	h.Logger.Fatalf(format, args...)
}

func (h *HybridLogger) Panic(args ...interface{}) { h.Logger.Panic(args...) }
func (h *HybridLogger) Panicf(format string, args ...interface{}) {
	h.Logger.Panicf(format, args...)
}
