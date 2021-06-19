package main

import (
	"fmt"
	"log"
	"time"

	"github.com/go-sql-driver/mysql"
	dbr "github.com/gocraft/dbr/v2"
)

func main() {
	log.Println("mysql orm of dbr run...")
	dbConf := &DBConf{
		Ip:           "127.0.0.1",
		Port:         3306,
		User:         "root",
		Password:     "root1234",
		Database:     "test",
		MaxIdleConns: 10,
		MaxOpenConns: 100,
		ParseTime:    true,
	}

	dsn, err := dbConf.DSN()
	if err != nil {
		log.Fatalln("err: ", err)
	}

	// opens a database
	conn, _ := dbr.Open("mysql", dsn, dbReceiver)

	// 设置连接池相关参数
	conn.SetMaxOpenConns(dbConf.MaxOpenConns)
	conn.SetMaxIdleConns(dbConf.MaxIdleConns)
	conn.SetConnMaxLifetime(dbConf.MaxLifetime)
	conn.SetConnMaxIdleTime(dbConf.MaxIdleTime)

	// 执行session查询操作
	users := make([]User, 0, 10)
	sess := handleSession(conn) // 每次查询都是一个session会话操作

	total, err := sess.Select("*").From(User{}.TableName()).Where("id > ?", 1).Load(&users)
	log.Println("total: ", total, "error: ", err)
	log.Println("users: ", users)
	// ... 其他方法可以参考 https://github.com/gocraft/dbr
}

// User
type User struct {
	Id   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Age  int    `json:"age" db:"age"`
}

// TableName table
func (User) TableName() string {
	return "user"
}

// DBConf DB config
type DBConf struct {
	Ip        string
	Port      int // 默认3306
	User      string
	Password  string
	Database  string
	Charset   string // 字符集 utf8mb4 支持表情符号
	Collation string // 整理字符集 utf8mb4_unicode_ci

	MaxIdleConns int // 空闲pool个数
	MaxOpenConns int // 最大open connection个数

	// sets the maximum amount of time a connection may be reused.
	// 设置连接可以重用的最大时间
	// 给db设置一个超时时间，时间小于数据库的超时时间
	MaxLifetime time.Duration // 数据库超时时间
	MaxIdleTime time.Duration // 最大空闲时间

	// 连接超时/读取超时/写入超时设置
	Timeout      time.Duration // Dial timeout
	ReadTimeout  time.Duration // I/O read timeout
	WriteTimeout time.Duration // I/O write timeout

	ParseTime bool   // 格式化时间类型
	Loc       string // 时区字符串 Local,PRC
}

func (conf *DBConf) DSN() (string, error) {
	if conf.Ip == "" {
		conf.Ip = "127.0.0.1"
	}

	if conf.Port == 0 {
		conf.Port = 3306
	}

	if conf.Charset == "" {
		conf.Charset = "utf8mb4"
	}

	// 默认字符序，定义了字符的比较规则
	if conf.Collation == "" {
		conf.Collation = "utf8mb4_general_ci"
	}

	if conf.Loc == "" {
		conf.Loc = "Local"
	}

	if conf.Timeout == 0 {
		conf.Timeout = 10 * time.Second
	}

	if conf.WriteTimeout == 0 {
		conf.WriteTimeout = 5 * time.Second
	}

	if conf.ReadTimeout == 0 {
		conf.ReadTimeout = 5 * time.Second
	}

	if conf.MaxLifetime == 0 {
		conf.MaxLifetime = 20 * time.Minute
	}

	if conf.MaxIdleTime == 0 {
		conf.MaxIdleTime = 10 * time.Minute
	}

	// mysql connection time loc.
	loc, err := time.LoadLocation(conf.Loc)
	if err != nil {
		return "", err
	}

	// mysql config
	mysqlConf := mysql.Config{
		User:   conf.User,
		Passwd: conf.Password,
		Net:    "tcp",
		Addr:   fmt.Sprintf("%s:%d", conf.Ip, conf.Port),
		DBName: conf.Database,
		// Connection parameters
		Params: map[string]string{
			"charset": conf.Charset,
		},
		Collation:            conf.Collation,
		Loc:                  loc,               // Location for time.Time values
		Timeout:              conf.Timeout,      // Dial timeout
		ReadTimeout:          conf.ReadTimeout,  // I/O read timeout
		WriteTimeout:         conf.WriteTimeout, // I/O write timeout
		AllowNativePasswords: true,              // Allows the native password authentication method
		ParseTime:            conf.ParseTime,    // Parse time values to time.Time
	}

	return mysqlConf.FormatDSN(), nil
}

func handleSession(con *dbr.Connection, timeout ...time.Duration) *dbr.Session {
	session := con.NewSession(nil)
	if len(timeout) > 0 && timeout[0] > 0 {
		session.Timeout = timeout[0] // 单个session会话超时处理
	}

	return session
}

// ===========对不同的事件进行监听================

var dbReceiver = &NullEventReceiver{}

// NullEventReceiver is a sentinel EventReceiver.
// Use it if the caller doesn't supply one.
type NullEventReceiver struct{}

// Event receives a simple notification when various events occur.
func (n *NullEventReceiver) Event(eventName string) {
	// log.Println("event_name: ", eventName)
}

// EventKv receives a notification when various events occur along with
// optional key/value data.
func (n *NullEventReceiver) EventKv(eventName string, kvs map[string]string) {
	// log.Println("event_name1: ", eventName)
	// log.Println("kvs: ", kvs)
}

// EventErr receives a notification of an error if one occurs.
func (n *NullEventReceiver) EventErr(eventName string, err error) error { return err }

// EventErrKv receives a notification of an error if one occurs along with
// optional key/value data.
func (n *NullEventReceiver) EventErrKv(eventName string, err error, kvs map[string]string) error {
	return err
}

// Timing receives the time an event took to happen.
func (n *NullEventReceiver) Timing(eventName string, nanoseconds int64) {
	log.Println("event_name: ", eventName)
	log.Println("nanoseconds: ", nanoseconds)
}

// TimingKv receives the time an event took to happen along with optional key/value data.
func (n *NullEventReceiver) TimingKv(eventName string, nanoseconds int64, kvs map[string]string) {
	log.Println("exec event_name: ", eventName)
	log.Println("nanoseconds: ", nanoseconds)
	log.Println("kvs: ", kvs)
}
