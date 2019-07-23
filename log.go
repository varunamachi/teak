package teak

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

//Level - gives log level
type Level int

const (
	//TraceLevel - low level debug message
	TraceLevel Level = 1

	//DebugLevel - a debug message
	DebugLevel Level = 2

	//InfoLevel - information message
	InfoLevel Level = 3

	//WarnLevel - warning message
	WarnLevel Level = 4

	//ErrorLevel - error message
	ErrorLevel Level = 5

	//FatalLevel - fatal messages
	FatalLevel Level = 6

	//PrintLevel - prints a output message
	PrintLevel Level = 7
)

//Writer - interface that takes a message and writes it based on
//the implementation
type Writer interface {
	UniqueID() string
	Write(message string)
	Enable(value bool)
	IsEnabled() (value bool)
}

//Logger - interface that defines a logger implementation
type Logger interface {
	//Log - logs a message with given level and module
	Log(level Level,
		module string,
		fmtstr string,
		args ...interface{})

	//RegisterWriter - registers a writer
	RegisterWriter(writer Writer)

	//RemoveWriter - removes a writer with given ID
	RemoveWriter(uniqueID string)

	//GetWriter - gives the writer with given ID
	GetWriter(uniqueID string) (writer Writer)
}

//LoggerConfig - configuration that is used to initialize the logger
type LoggerConfig struct {
	Logger      Logger
	LogConsole  bool
	FilterLevel Level
}

var lconf = LoggerConfig{
	Logger:      NewDirectLogger(),
	LogConsole:  false,
	FilterLevel: InfoLevel,
}

//ToString - maps level to a string
func ToString(level Level) string {
	switch level {
	case TraceLevel:
		return "[TRACE]"
	case DebugLevel:
		return "[DEBUG]"
	case InfoLevel:
		return "[ INFO]"
	case WarnLevel:
		return "[ WARN]"
	case ErrorLevel:
		return "[ERROR]"
	case FatalLevel:
		return "[FATAL]"
	}
	return "[     ]"
}

func (level Level) String() string {
	return ToString(level)
}

//InitLogger - initializes the logger with non default options. If you
//want default behavior, no need to call any init functions
func InitLogger(lc LoggerConfig) {
	lconf = lc
	if lc.LogConsole {
		lconf.Logger.RegisterWriter(NewConsoleWriter())
	}
}

//SetLevel - sets the filter level
func SetLevel(level Level) {
	lconf.FilterLevel = level
}

//GetLevel - gets the filter level
func GetLevel() (level Level) {
	return lconf.FilterLevel
}

//Trace - trace logs
func Trace(module, fmtStr string, args ...interface{}) {
	if TraceLevel >= lconf.FilterLevel {
		lconf.Logger.Log(TraceLevel, module, fmtStr, args...)
	}
}

//Debug - debug logs
func Debug(module, fmtStr string, args ...interface{}) {
	if DebugLevel >= lconf.FilterLevel {
		lconf.Logger.Log(DebugLevel, module, fmtStr, args...)
	}
}

//Info - information logs
func Info(module, fmtStr string, args ...interface{}) {
	if InfoLevel >= lconf.FilterLevel {
		lconf.Logger.Log(InfoLevel, module, fmtStr, args...)
	}
}

//Warn - warning logs
func Warn(module, fmtStr string, args ...interface{}) {
	if WarnLevel >= lconf.FilterLevel {
		lconf.Logger.Log(WarnLevel, module, fmtStr, args...)
	}
}

//Error - error logs
func Error(module, fmtStr string, args ...interface{}) {
	if ErrorLevel >= lconf.FilterLevel {
		lconf.Logger.Log(ErrorLevel, module, fmtStr, args...)
		// Print(module, fmtStr, args...)
	}
}

//Fatal - error logs
func Fatal(module, fmtStr string, args ...interface{}) {
	lconf.Logger.Log(FatalLevel, module, fmtStr, args...)
	// Print(module, fmtStr, args...)
	os.Exit(-1)
}

//LogError - error log
func LogError(module string, err error) error {
	if err != nil && ErrorLevel >= lconf.FilterLevel {
		_, file, line, _ := runtime.Caller(1)
		lconf.Logger.Log(ErrorLevel, module, "%s -- %s @ %d",
			err.Error(),
			file,
			line)
		// LogJSON(ErrorLevel, module, err)
	}
	return err
}

//LogErrorX - log error with a message
func LogErrorX(module, msg string, err error) error {
	if err != nil && ErrorLevel >= lconf.FilterLevel {
		_, file, line, _ := runtime.Caller(1)
		lconf.Logger.Log(ErrorLevel, module, "%s -- %s. ERROR: %s @ %d",
			msg,
			err.Error(),
			file,
			line)
		// LogJSON(ErrorLevel, module, err)
	}
	return err
}

//LogFatal - logs before exit
func LogFatal(module string, err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		lconf.Logger.Log(FatalLevel, module, "%v -- %s @ %d", err, file, line)
		// Print(module, "%v", err)
		os.Exit(-1)
	}
}

//Print - prints the message on console
func Print(module, fmtStr string, args ...interface{}) {
	lconf.Logger.Log(PrintLevel, module, fmtStr, args)
	fmt.Printf(fmtStr+"\n", args...)
}

//LogJSON - logs data as JSON
func LogJSON(level Level, module string, data interface{}) {
	b, err := json.MarshalIndent(data, "", "    ")
	if err == nil {
		lconf.Logger.Log(level, module, "%s", string(b))
	}
}

//HasError - logs the errors from the array that are not nil and return true if
//there were one or more non nil errors
func HasError(module string, errs ...error) (has bool) {
	_, file, line, _ := runtime.Caller(1)
	for _, e := range errs {
		if e != nil {
			if ErrorLevel >= lconf.FilterLevel {
				lconf.Logger.Log(ErrorLevel, module, "%s -- %s @ %d",
					e.Error(),
					file,
					line)
			}
			has = true
		}
	}
	return has
}

//DirectLogger - logger that writes directly to all registered writers
type DirectLogger struct {
	writers map[string]Writer
}

//NewDirectLogger - creates a new DirectLogger instace
func NewDirectLogger() *DirectLogger {
	return &DirectLogger{
		writers: make(map[string]Writer),
	}
}

//Log - logs a message with given level and module
func (dl *DirectLogger) Log(level Level,
	module string,
	fmtstr string,
	args ...interface{}) {
	if level == PrintLevel {
		return
	}
	fmtstr = ToString(level) + " [" + module + "] " + fmtstr
	msg := fmt.Sprintf(fmtstr, args...)
	for _, writer := range dl.writers {
		if writer.IsEnabled() {
			writer.Write(msg)
		}
	}
}

//RegisterWriter - registers a writer
func (dl *DirectLogger) RegisterWriter(writer Writer) {
	if writer != nil {
		dl.writers[writer.UniqueID()] = writer
	}
}

//RemoveWriter - removes a writer with given ID
func (dl *DirectLogger) RemoveWriter(uniqueID string) {
	delete(dl.writers, uniqueID)
}

//GetWriter - gives the writer with given ID
func (dl *DirectLogger) GetWriter(uniqueID string) (writer Writer) {
	return dl.writers[uniqueID]
}

//AsyncLogger - logger that uses goroutine for dispatching
type AsyncLogger struct {
	sync.Mutex
	writers map[string]Writer
}

//NewAsyncLogger - creates a new DirectLogger instace
func NewAsyncLogger() *AsyncLogger {
	return &AsyncLogger{
		writers: make(map[string]Writer),
	}
}

//Log - logs a message with given level and module
func (al *AsyncLogger) Log(level Level,
	module string,
	fmtstr string,
	args ...interface{}) {
	if level == PrintLevel {
		return
	}
	go func() {
		fmtstr = ToString(level) + " [" + module + "] " + fmtstr
		msg := fmt.Sprintf(fmtstr, args...)
		al.Lock()
		for _, writer := range al.writers {
			if writer.IsEnabled() {
				writer.Write(msg)
			}
		}
		al.Unlock()
	}()
}

//RegisterWriter - registers a writer
func (al *AsyncLogger) RegisterWriter(writer Writer) {
	if writer != nil {
		al.Lock()
		al.writers[writer.UniqueID()] = writer
		al.Unlock()
	}
}

//RemoveWriter - removes a writer with given ID
func (al *AsyncLogger) RemoveWriter(uniqueID string) {
	al.Lock()
	delete(al.writers, uniqueID)
	al.Unlock()
}

//GetWriter - gives the writer with given ID
func (al *AsyncLogger) GetWriter(uniqueID string) (writer Writer) {
	al.Lock()
	l := al.writers[uniqueID]
	al.Unlock()
	return l
}

//ConsoleWriter - Log writer that writes to console
type ConsoleWriter struct {
	enabled bool
}

//NewConsoleWriter - creates a new console writer
func NewConsoleWriter() *ConsoleWriter {
	return &ConsoleWriter{
		enabled: true,
	}
}

//UniqueID - identifier for console writer
func (cw *ConsoleWriter) UniqueID() string {
	return "console"
}

//Write - writes message to console
func (cw *ConsoleWriter) Write(message string) {
	if cw.enabled {
		fmt.Println(message)
	}
}

//Enable - enables or disables console logger based on the passed value
func (cw *ConsoleWriter) Enable(value bool) {
	cw.enabled = value
}

//IsEnabled - tells if the writer is enabled
func (cw *ConsoleWriter) IsEnabled() (value bool) {
	return cw.enabled
}

//Event - represents a event initiated by a user while performing an operation
type Event struct {
	Op       string      `json:"op" bson:"op"`
	UserID   string      `json:"userID" bson:"userID"`
	UserName string      `json:"userName" bson:"userName"`
	Success  bool        `json:"success" bson:"success"`
	Error    string      `json:"error" bson:"error"`
	Time     time.Time   `json:"time" bson:"time"`
	Data     interface{} `json:"data" bson:"data"`
}

//EventAuditor - handles application events for audit purposes
type EventAuditor interface {
	//LogEvent - logs given event into storage
	LogEvent(event *Event)

	//GetEvents - retrieves event entries based on filters
	GetEvents(offset, limit int,
		filter *Filter) (
		total int,
		events []*Event,
		err error)

	//CreateIndices - creates mongoDB indeces for tables used for event logs
	CreateIndices() (err error)

	//CleanData - cleans event related data from database
	CleanData() (err error)
}

//NoOpAuditor - doesnt do anything, it's a dummy auditor
type NoOpAuditor struct{}

//LogEvent - logs event to console
func (n *NoOpAuditor) LogEvent(event *Event) {
	if event.Success {
		fmt.Printf("Event:Info - %s BY %s", event.Op, event.UserID)
	} else {
		fmt.Printf("Event:Error - %s BY %s", event.Op, event.UserID)
	}
}

//GetEvents - gives an empty list of events
func (n *NoOpAuditor) GetEvents(
	offset, limit int, filter *Filter) (
	total int, events []*Event, err error) {
	return total, events, err
}

//CreateIndices - creates nothing
func (n *NoOpAuditor) CreateIndices() (err error) { return err }

//CleanData - there's nothing to clean
func (n *NoOpAuditor) CleanData() (err error) { return err }

var eventAuditor EventAuditor

//SetEventAuditor - sets the event auditor
func SetEventAuditor(auditor EventAuditor) {
	eventAuditor = auditor
}

//GetAuditor - gets the event auditor
func GetAuditor() EventAuditor {
	return eventAuditor
}

//LogEvent - logs an event using the registered audit function
func LogEvent(
	op string,
	userID string,
	userName string,
	success bool,
	err string,
	data interface{}) {
	eventAuditor.LogEvent(&Event{
		Op:       op,
		UserID:   userID,
		UserName: userName,
		Success:  success,
		Error:    err,
		Time:     time.Now(),
		Data:     data,
	})
}
