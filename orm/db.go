package orm

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/juju/errors"
	//mysql
	_ "github.com/go-sql-driver/mysql"
)

// DB db instance.
type DB struct {
	IP       string
	Port     int
	DBName   string
	UserName string
	Passwd   string
	Charset  string
	Timeout  int
}

// NewDB create db instance, timeout 单位:秒.
func NewDB(ip string, port int, dbName, user, pass, charset string, timeout int) *DB {
	return &DB{
		IP:       ip,
		Port:     port,
		DBName:   dbName,
		UserName: user,
		Passwd:   pass,
		Charset:  charset,
		Timeout:  timeout,
	}
}

// GetConnection open new connect to db.
func (db *DB) GetConnection() (*sql.DB, error) {
	dsn := db.getDSN()
	stmtDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, errors.Trace(err)
	}

	stmtDB.SetMaxOpenConns(0)
	if err := stmtDB.Ping(); err != nil {
		return nil, errors.Trace(err)
	}

	return stmtDB, nil
}

const (
	maxAllowedPacket = 134217728
)

func (db *DB) getDSN() string {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", db.UserName, db.Passwd, db.IP, db.Port, db.DBName)

	if optStr := db.getOpt(); optStr != "" {
		dsn = fmt.Sprintf("%s?%s", dsn, optStr)
	}

	return dsn
}

func (db *DB) getOpt() string {
	var opts []string

	if len(db.Charset) > 0 {
		opts = append(opts, fmt.Sprintf("charset=%s", db.Charset))
	}

	if db.Timeout > 0 {
		opts = append(opts, fmt.Sprintf("timeout=%ds", db.Timeout))
	}

	opts = append(opts, fmt.Sprintf("maxAllowedPacket=%d", maxAllowedPacket))

	return strings.Join(opts, "&")
}
