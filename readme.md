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
         1. Expanded on the basis of database/sql, adding convenient methods such as select, insert, delete, and update;
         2. Support session ctx binding, transaction tx ctx binding, global session binding, and single session binding;
         3. The standard sql.DB Exec method is also supported, because *sql.DB is nested inside dbr,
            so that sql native methods can be executed directly;
         4. dbr supports opentracing monitoring, provides an interface design mode, 
            and can perform performance monitoring and analysis on each operation of the database;
         5. Developers can get started quickly, and they don't need to learn some complex syntax of ORM components,
            nor do they need to understand the underlying implementation details of ORM, and there is no mental burden;
         6. Developers only need to master the mysql syntax to quickly develop, and the operation is 
            relatively simple and intuitive;

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
```

