# dbr demo
    
    mysql dbr demo

# ref
    
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
            
# Core code implementation
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
# event core code

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
