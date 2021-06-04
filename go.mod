module github.com/topxeq/umx

go 1.16

require (
	github.com/go-sql-driver/mysql v1.5.0
	github.com/kardianos/service v1.2.0
	github.com/mattn/go-sqlite3 v1.14.6
	github.com/topxeq/sqltk v0.0.0
	github.com/topxeq/tk v0.0.0
)

replace github.com/topxeq/tk v0.0.0 => ../tk

replace github.com/topxeq/sqltk v0.0.0 => ../sqltk
