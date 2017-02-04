package orm

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

func TestORMStruct(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.passwd from userinfo limit 1"
	result := struct {
		ID       int64  `db:"id"`
		User     string `db:"user"`
		Password string `db:"passwd"`
	}{}
	sql, _, err := SqlQuery(&result, queryTable("userinfo"))
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s, recv:%s", expect, sql)
	}
}

func TestORMArray(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.passwd from userinfo"
	result := []struct {
		ID       int64  `db:"id"`
		User     string `db:"user"`
		Password string `db:"passwd"`
	}{}
	sql, _, err := SqlQuery(&result, queryTable("userinfo"))
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s, recv:%s", expect, sql)
	}
}

func TestORMSort(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.passwd from userinfo order by user"
	result := []struct {
		ID       int64  `db:"id"`
		User     string `db:"user"`
		Password string `db:"passwd"`
	}{}
	sql, _, err := SqlQuery(&result, queryTable("userinfo"), QuerySort("user"))
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s, recv:%s", expect, sql)
	}
}

func TestORMSortOrder(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.passwd from userinfo order by user desc"
	result := []struct {
		ID       int64  `db:"id"`
		User     string `db:"user"`
		Password string `db:"passwd"`
	}{}
	sql, _, err := SqlQuery(&result, queryTable("userinfo"), QuerySort("user"), QueryOrder("desc"))
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s, recv:%s", expect, sql)
	}
}

func TestORMLimit(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.passwd from userinfo limit 10"
	result := []struct {
		ID       int64  `db:"id"`
		User     string `db:"user"`
		Password string `db:"passwd"`
	}{}
	sql, _, err := SqlQuery(&result, queryTable("userinfo"), QueryLimit(10))
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s, recv:%s", expect, sql)
	}
}

func TestORMLimitOffset(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.passwd from userinfo limit 5,10"
	result := []struct {
		ID       int64  `db:"id"`
		User     string `db:"user"`
		Password string `db:"passwd"`
	}{}
	sql, _, err := SqlQuery(&result, queryTable("userinfo"), QueryLimit(10), QueryOffset(5))
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s, recv:%s", expect, sql)
	}
}

func TestORMWhere(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.passwd from userinfo where id=1010"
	result := []struct {
		ID       int64  `db:"id"`
		User     string `db:"user"`
		Password string `db:"passwd"`
	}{}
	sql, _, err := SqlQuery(&result, queryTable("userinfo"), QueryWhere("id=1010"))
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s, recv:%s", expect, sql)
	}
}

func TestORMMutilTalbe(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.passwd, ext.qq from userinfo, ext where ext.user_id=userinfo.id and id=1010"
	result := []struct {
		ID       int64  `db:"id"`
		User     string `db:"user"`
		Password string `db:"passwd"`
		QQ       string `db:"ext.qq"`
	}{}
	sql, _, err := SqlQuery(&result, queryTable("userinfo, ext"), QueryWhere("ext.user_id=userinfo.id and id=1010"))
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s, recv:%s", expect, sql)
	}
}

