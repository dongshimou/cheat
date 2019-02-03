package logger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const VERSION = "2.0.3"

const (
	ErrorNotInit = "logger未初始化"
)

func init() {
	defaultLogger = New("")
	log.Println("logger VERSION: ", VERSION)
	log.SetFlags(log.Lshortfile)
}

//flag常量
const (
	Ldate = 1 << iota
	Ltime
	Lmicroseconds
	Llongfile
	Lshortfile
	Llevel
	LstdFlags = Ldate | Ltime                   //提供一些基础的设置
	Ldefault  = Llevel | Lshortfile | LstdFlags //默认设置
)

//等级
const (
	LEVEL_INFO     = "INFO"     //消息在粗粒度级别上突出强调应用程序的运行过程
	LEVEL_REQUEST  = "REQUEST"  //http请求的日志
	LEVEL_DEBUG    = "DEBUG"    //细粒度信息事件对调试应用程序是非常有帮助的
	LEVEL_WARNNING = "WARNNING" //潜在错误的情形
	LEVEL_ERROR    = "ERROR"    //虽然发生错误事件，但仍然不影响系统的继续运行
	LEVEL_BEHAVIOR = "BEHAVIOR" //行为日志
)

var levelMap = map[string]int{
	LEVEL_DEBUG:    1,
	LEVEL_REQUEST:  2,
	LEVEL_INFO:     3,
	LEVEL_BEHAVIOR: 4,
	LEVEL_WARNNING: 5,
	LEVEL_ERROR:    6,
}

func GetLevel(label string) int {
	val, ok := levelMap[label]
	if ok {
		return val
	}
	return 0
}

//输出
const (
	STD = iota
	FILE
)

type Logger struct {
	mu      sync.Mutex   //保护其他字段
	buf     bytes.Buffer //封装了一些操作[]byte的方法，用起来更方便
	out     io.Writer    //输出到终端
	level   int
	service string
	//File    *LogFile //输出到文件,fd永远指向当天文件
}

type LogRequestStruct struct {
	LogEntryStruct
	TimeStr    string `json:"time_str"`
	Path       string `json:"path"`
	Host       string `json:"host"`
	Method     string `json:"method"`
	RemoteAddr string `json:"remote_addr"`
	UserAgent  string `json:"user_agent"`
	DeviceId   string `json:"device_id"`
	Referrer   string `json:"referrer"`
	Service    string `json:"service"`
}

type LogEntryStruct struct {
	Topic   string `json:"topic"`
	Time    int64  `json:"time"`
	Service string `json:"service"`
	UserID  int64  `json:"user_id"`
	Msg     string `json:"msg"` //可以是错误信息 也可以 是 提示信息
}

var loggerMap = map[string]*Logger{}

var defaultLogger *Logger

func GetLogger(service string) *Logger {
	logger, ok := loggerMap[service]
	if ok == false {
		return nil
	}
	return logger
}

func GetDefaultLogger() *Logger {
	return defaultLogger
}

func New(service string) *Logger {

	log.Println(" New logger", service)
	exist, ok := loggerMap[service]
	if ok {
		return exist
	}
	logger := &Logger{
		out:     os.Stdout,
		service: service,
	}
	loggerMap[service] = logger
	if (defaultLogger == nil) || defaultLogger.service == "" {
		defaultLogger = logger
	}
	return logger
}

func SetDefaultLogger(service string) *Logger {
	logger, ok := loggerMap[service]
	if ok {
		defaultLogger = logger
		return defaultLogger
	}
	log.Fatal("set detault logger", service)
	return nil
}

func itoa(buf *bytes.Buffer, i int, wid int) {
	var u uint = uint(i)

	if u == 0 && wid <= 1 {
		buf.WriteByte('0')
		return
	}

	var b [32]byte
	bp := len(b)
	for ; u > 0 || wid > 0; u /= 10 {
		bp--
		wid--
		b[bp] = byte(u%10) + '0'
	}

	for bp < len(b) {
		buf.WriteByte(b[bp])
		bp++
	}
}

func structsToMap(a interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(a)

	data := map[string]interface{}{}

	err = json.Unmarshal(b, &data)
	if err != nil {
		return data, err
	}

	return data, nil

}

func InterfaceJoin(msg ...interface{}) string {
	s := []string{}
	for _, i := range msg {
		s = append(s, fmt.Sprintf("%v", i))
	}
	return strings.Join(s, " ")
}

func (l *Logger) SetLevel(levelName string) {
	l.level = GetLevel(levelName)
	log.Println(" set logger level ", levelName, l.level)
}

func (l *Logger) writeLine(level string, depth int, msg ...interface{}) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var file string
	var line int
	var ok bool
	_, file, line, ok = runtime.Caller(depth)
	locate := ""

	if !ok {
		file = ""
		line = 0
	} else {
		locate = file + ":" + strconv.Itoa(line)
	}

	lineEntry := []byte(time.Now().Format("2006-01-02 15:04:05") + " " + level + " " + locate + " " + InterfaceJoin(msg...) + "\n")

	var err error
	if GetLevel(level) >= l.level {
		_, err = l.out.Write(lineEntry)
	}

	//_, err = l.File.Write(lineEntry)

	if err != nil {
		log.Println(" logger writeLine ERROR ", err)
	}

	return err
}

func (l *Logger) logEntry(topic string, entry interface{}) error {

	logBytes, err := json.Marshal(entry)
	if err != nil {
		log.Println("ERROR", err)
		return err
	}
	return l.writeLine(LEVEL_BEHAVIOR, 3, topic, string(logBytes))
}

func (l *Logger) logRequest(topic string, entry interface{}) error {

	logBytes, err := json.Marshal(entry)
	if err != nil {
		log.Println("ERROR", err)
		return err
	}
	return l.writeLine(LEVEL_REQUEST, 3, topic, string(logBytes))
}

func Debug(msg ...interface{}) error {
	if defaultLogger == nil {
		log.Println(msg...)
		return errors.New(ErrorNotInit)
	}
	return defaultLogger.writeLine(LEVEL_DEBUG, 2, msg...)
}

func Debugf(format string, v ...interface{}) error {
	if defaultLogger == nil {
		log.Printf(format, v...)
		return errors.New(ErrorNotInit)
	}

	return defaultLogger.writeLine(LEVEL_DEBUG, 2, fmt.Sprintf(format, v...))
}

func Info(msg ...interface{}) error {
	if defaultLogger == nil {
		log.Println(msg...)
		return errors.New(ErrorNotInit)
	}
	return defaultLogger.writeLine(LEVEL_INFO, 2, msg...)
}

func Behavior(topic string, userID int64, data map[string]interface{}) error {
	if defaultLogger == nil {
		log.Println(topic, data)
		return errors.New(ErrorNotInit)
	}
	entry := LogEntryStruct{}
	entry.Time = time.Now().UnixNano() / 1e6
	entry.Topic = topic
	entry.UserID = userID

	entry.Service = defaultLogger.service
	m, _ := structsToMap(entry)

	for k, v := range data {
		_, ok := m[k]
		if ok == true { //跳过关键字
			continue
		} else {
			m[k] = v
		}
	}
	return defaultLogger.logEntry(topic, m)
}

func Error(msg ...interface{}) error {
	if defaultLogger == nil {
		log.Println(msg...)
		return errors.New(ErrorNotInit)
	}
	return defaultLogger.writeLine(LEVEL_ERROR, 2, msg...)
}
func Warnning(msg ...interface{}) error {
	if defaultLogger == nil {
		log.Println(msg...)
		return errors.New(ErrorNotInit)
	}
	return defaultLogger.writeLine(LEVEL_WARNNING, 2, msg...)
}

func (l *Logger) LogDebug(msg ...interface{}) error {
	return l.writeLine(LEVEL_DEBUG, 2, msg...)
}
func (l *Logger) LogInfo(topic string, data map[string]interface{}) error {
	entry := LogEntryStruct{}
	entry.Time = time.Now().UnixNano() / 1e6
	entry.Topic = topic

	entry.Service = l.service
	m, _ := structsToMap(entry)

	for k, v := range data {
		_, ok := m[k]
		if ok == true { //跳过关键字
			continue
		} else {
			m[k] = v
		}
	}
	return l.logEntry(topic, m)
}
func (l *Logger) LogError(msg ...interface{}) error {
	return l.writeLine(LEVEL_ERROR, 2, msg...)
}
func (l *Logger) LogWarnning(msg ...interface{}) error {
	return l.writeLine(LEVEL_WARNNING, 2, msg...)
}
