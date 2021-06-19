# dbr demo
    
    mysql dbr demo

# dbr design ideas
    
    https://github.com/gocraft/dbr
    dbr优秀的设计体现如下：
        1、在database/sql基础上进行了拓展,增加了select,insert,delete,update等便捷的方法；
        2、支持session ctx绑定，同时也支持事务tx ctx绑定，支持全局session绑定，也支持单个session绑定；
        3、对标准的sql.DB Exec方法也是支持的，因为dbr内部嵌套了*sql.DB,这样可以直接执行sql原生方法；
        4、dbr支持opentracing打点监控，提供了接口设计模式，可以对数据库每个操作做性能监控分析；
        5、开发人员既能快速上手，又不需要学习一些orm组件的复杂语法，也不必了解ORM底层实现细节，没有心智负担；
        6、开发人员只需要掌握mysql语法就可以快速进行开发，操作起来比较简单、直观；
    The excellent design of dbr is as follows:
         1. Expanded on the basis of database/sql, adding convenient methods such as select, insert,
            delete, and update;
         2. Support session ctx binding, transaction tx ctx binding, global session binding, 
            and single session binding;
         3. The standard sql.DB Exec method is also supported, because *sql.DB is nested inside dbr,
            so that sql native methods can be executed directly;
         4. dbr supports opentracing monitoring, provides an interface design mode, 
            and can perform performance monitoring and analysis on each operation
            of the database;
         5. Developers can get started quickly, and they don't need to learn some complex syntax 
            of ORM components,nor do they need to understand the underlying implementation details
            of ORM, and there is no mental burden;
         6. Developers only need to master the mysql syntax to quickly develop, and the operation is 
            relatively simple and intuitive;
# godoc
	
https://pkg.go.dev/github.com/gocraft/dbr#readme-examples

# Quick to use 

``` go
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
```

# dbr core code implementation
``` go
// Connection wraps sql.DB with an EventReceiver
// to send events, errors, and timings.
type Connection struct {
    *sql.DB
    Dialect
    EventReceiver
}

// Session represents a business unit of execution.
//
// All queries in gocraft/dbr are made in the context of a session.
// This is because when instrumenting your app, it's important
// to understand which business action the query took place in.
//
// A custom EventReceiver can be set.
//
// Timeout specifies max duration for an operation like Select.
type Session struct {
    *Connection
    EventReceiver
    Timeout time.Duration
}

// SessionRunner can do anything that a Session can except start a transaction.
// Both Session and Tx implements this interface.
type SessionRunner interface {
	Select(column ...string) *SelectBuilder
	SelectBySql(query string, value ...interface{}) *SelectBuilder

	InsertInto(table string) *InsertBuilder
	InsertBySql(query string, value ...interface{}) *InsertBuilder

	Update(table string) *UpdateBuilder
	UpdateBySql(query string, value ...interface{}) *UpdateBuilder

	DeleteFrom(table string) *DeleteBuilder
	DeleteBySql(query string, value ...interface{}) *DeleteBuilder
}

type runner interface {
	GetTimeout() time.Duration
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

// dbr interface func mode
package dbr

// Builder builds SQL in Dialect like MySQL, and PostgreSQL.
// The raw SQL and values are stored in Buffer.
//
// The core of gocraft/dbr is interpolation, which can expand ? with arbitrary SQL.
// If you need a feature that is not currently supported, you can build it
// on your own (or use Expr).
//
// To do that, the value that you wish to be expanded with ? needs to
// implement Builder.
type Builder interface {
	Build(Dialect, Buffer) error
}

// BuildFunc implements Builder.
type BuildFunc func(Dialect, Buffer) error

// Build calls itself to build SQL.
func (b BuildFunc) Build(d Dialect, buf Buffer) error {
	return b(d, buf)
}
```
# dbr event core code

``` go
// EventReceiver gets events from dbr methods for logging purposes.
type EventReceiver interface {
	Event(eventName string)
	EventKv(eventName string, kvs map[string]string)
	EventErr(eventName string, err error) error
	EventErrKv(eventName string, err error, kvs map[string]string) error
	Timing(eventName string, nanoseconds int64)
	TimingKv(eventName string, nanoseconds int64, kvs map[string]string)
}

// TracingEventReceiver is an optional interface an EventReceiver type can implement
// to allow tracing instrumentation
type TracingEventReceiver interface {
	SpanStart(ctx context.Context, eventName, query string) context.Context
	SpanError(ctx context.Context, err error)
	SpanFinish(ctx context.Context)
}

// opentracing/event_receiver.go 
package opentracing

import (
	"context"

	ot "github.com/opentracing/opentracing-go"
	otext "github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log"
)

// EventReceiver provides an embeddable implementation of dbr.TracingEventReceiver
// powered by opentracing-go.
type EventReceiver struct{}

// SpanStart starts a new query span from ctx, then returns a new context with the new span.
func (EventReceiver) SpanStart(ctx context.Context, eventName, query string) context.Context {
	span, ctx := ot.StartSpanFromContext(ctx, eventName)
	otext.DBStatement.Set(span, query)
	otext.DBType.Set(span, "sql")
	return ctx
}

// SpanFinish finishes the span associated with ctx.
func (EventReceiver) SpanFinish(ctx context.Context) {
	if span := ot.SpanFromContext(ctx); span != nil {
		span.Finish()
	}
}

// SpanError adds an error to the span associated with ctx.
func (EventReceiver) SpanError(ctx context.Context, err error) {
	if span := ot.SpanFromContext(ctx); span != nil {
		otext.Error.Set(span, true)
		span.LogFields(otlog.String("event", "error"), otlog.Error(err))
	}
}
```
