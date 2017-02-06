package orm

import (
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	dbc *DB
)

func init() {
	/*
		   CREATE TABLE `userinfo` (
		   `id` bigint(20) unsigned NOT NULL AUTO_INCREMENT,
		   `user` varchar(32) DEFAULT NULL,
		   `password` varchar(32) DEFAULT NULL,
		   `ctime` timestamp NOT NULL DEFAULT '0000-00-00 00:00:00',
		   `mtime` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		   PRIMARY KEY (`id`)
		   ) ENGINE=MyISAM;

		dbc = NewDB("127.0.0.1", 3306, "test", "orm_test", "orm_test_password", "utf8", "10")
	*/
}

func TestORMStruct(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.password from userinfo limit 1"
	result := struct {
		ID       int64
		User     string
		Password string
	}{}
	sql, err := NewStmt(nil, "userinfo").SQLQueryBuilder(&result)
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s,\n recv:%s.", expect, sql)
	}
}

func TestORMArray(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.password from userinfo"
	result := []struct {
		ID       int64
		User     string
		Password string
	}{}
	s := NewStmt(nil, "userinfo")
	sql, err := s.SQLQueryBuilder(&result)
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s, \nrecv:%s.", expect, sql)
	}
}

func TestORMSort(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.password from userinfo order by user"
	result := []struct {
		ID       int64
		User     string
		Password string
	}{}
	sql, err := NewStmt(nil, "userinfo").Sort("user").SQLQueryBuilder(&result)
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s, \nrecv:%s.", expect, sql)
	}
}

func TestORMSortOrder(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.password from userinfo order by user desc"
	result := []struct {
		ID       int64
		User     string
		Password string
	}{}
	sql, err := NewStmt(nil, "userinfo").Sort("user").Order("desc").SQLQueryBuilder(&result)
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s, \nrecv:%s.", expect, sql)
	}
}

func TestORMLimit(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.password from userinfo limit 10"
	result := []struct {
		ID       int64
		User     string
		Password string
	}{}
	sql, err := NewStmt(nil, "userinfo").Limit(10).SQLQueryBuilder(&result)
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s, \nrecv:%s.", expect, sql)
	}
}

func TestORMLimitOffset(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.password from userinfo limit 5,10"
	result := []struct {
		ID       int64
		User     string
		Password string
	}{}
	sql, err := NewStmt(nil, "userinfo").Limit(10).Offset(5).SQLQueryBuilder(&result)
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s, \nrecv:%s.", expect, sql)
	}
}

func TestORMWhere(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.password from userinfo where id=1010"
	result := []struct {
		ID       int64
		User     string
		Password string
	}{}
	sql, err := NewStmt(nil, "userinfo").Where("id=1010").SQLQueryBuilder(&result)
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s, \nrecv:%s.", expect, sql)
	}
}

func TestORMMutilTalbe(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.password, ext.qq from userinfo, ext where ext.user_id=userinfo.id and id=1010"
	result := []struct {
		ID       int64
		User     string
		Password string
		QQ       string `db:"ext.qq"`
	}{}
	sql, err := NewStmt(nil, "userinfo, ext").Where("ext.user_id=userinfo.id and id=1010").SQLQueryBuilder(&result)
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s, \nrecv:%s.", expect, sql)
	}
}

func TestORMQuerySlice(t *testing.T) {
	result := []struct {
		ID       int64
		User     string
		Password string
	}{}

	if dbc == nil {
		return
	}

	db, err := dbc.GetConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer db.Close()

	if err = NewStmt(db, "userinfo").Sort("id").Order("desc").Query(&result); err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("result:%+v", result)
}

func TestORMQueryOne(t *testing.T) {
	result := struct {
		ID       int64
		User     string
		Password string
	}{}

	if dbc == nil {
		return
	}

	db, err := dbc.GetConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer db.Close()

	if err = NewStmt(db, "userinfo").Where("id=2").Query(&result); err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("result:%+v", result)
}

func TestORMUpdate(t *testing.T) {
	data := struct {
		User     string
		Password string
	}{
		User:     fmt.Sprintf("new_user_%d", time.Now().Unix()),
		Password: fmt.Sprintf("new_password_%d", time.Now().Unix()),
	}
	if dbc == nil {
		return
	}

	db, err := dbc.GetConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer db.Close()

	id, err := NewStmt(db, "userinfo").Where("id=2").Update(&data)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("id:%+v", id)
}

func TestORMInsert(t *testing.T) {
	data := struct {
		ID       int64 `db_defult:"auto"`
		User     string
		Password string
	}{
		User:     fmt.Sprintf("user_%d", time.Now().Unix()),
		Password: fmt.Sprintf("password_%d", time.Now().Unix()),
	}

	if dbc == nil {
		return
	}

	db, err := dbc.GetConnection()
	if err != nil {
		t.Fatal(err.Error())
	}
	defer db.Close()

	id, err := NewStmt(db, "userinfo").Insert(&data)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("id:%+v", id)
}
