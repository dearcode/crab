package orm

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/dearcode/crab/log"
)

func TestORMStructDistinct(t *testing.T) {
	expect := "select userinfo.id, distinct userinfo.user, avg(age), sum(size), userinfo.password from userinfo limit 1"
	result := struct {
		ID       int64
		User     string `db:"user,distinct"`
		Age      int    `db:"avg(age)"`
		Size     int    `db:"sum(size)"`
		Password string
	}{}
	sql, err := NewStmt(nil, "userinfo").sqlQueryBuilder(&result)
	if err != nil {
		t.Fatal(err.Error())
	}
	if sql != expect {
		t.Fatalf("expect:%s,\n recv:%s.", expect, sql)
	}
}

func TestORMStruct(t *testing.T) {
	expect := "select userinfo.id, userinfo.user, userinfo.password from userinfo limit 1"
	result := struct {
		ID       int64
		User     string
		Password string
	}{}
	sql, err := NewStmt(nil, "userinfo").sqlQueryBuilder(&result)
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
	sql, err := s.sqlQueryBuilder(&result)
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
	sql, err := NewStmt(nil, "userinfo").Sort("user").sqlQueryBuilder(&result)
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
	sql, err := NewStmt(nil, "userinfo").Sort("user").Order("desc").sqlQueryBuilder(&result)
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
	sql, err := NewStmt(nil, "userinfo").Limit(10).sqlQueryBuilder(&result)
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
	sql, err := NewStmt(nil, "userinfo").Limit(10).Offset(5).sqlQueryBuilder(&result)
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
	sql, err := NewStmt(nil, "userinfo").Where("id=1010").sqlQueryBuilder(&result)
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
	sql, err := NewStmt(nil, "userinfo, ext").Where("ext.user_id=userinfo.id and id=1010").sqlQueryBuilder(&result)
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

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("select userinfo.id, userinfo.user, userinfo.password from userinfo order by id desc").WillReturnRows(sqlmock.NewRows([]string{"id", "user", "password"}).AddRow(3, "333", "3333").AddRow(1, "111", "1111"))

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

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectQuery("select userinfo.id, userinfo.user, userinfo.password from userinfo where id=(.+) limit 1").WillReturnRows(sqlmock.NewRows([]string{"id", "user", "password"}).AddRow(2, "tgy", "123456"))

	if err = NewStmt(db, "userinfo").Where("id=2").Query(&result); err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("result:%+v", result)
}

func TestORMUpdate(t *testing.T) {
	data := struct {
		User       string
		Password   string
		UpdateTime string `db:"utime" db_const:"now()"`
	}{
		User:     fmt.Sprintf("new_user_%d", time.Now().Unix()),
		Password: fmt.Sprintf("new_password_%d", time.Now().Unix()),
	}

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("update `userinfo` set `user`=(.+), `password`=(.+) where id=(.+)").WithArgs(data.User, data.Password).WillReturnResult(sqlmock.NewResult(0, 1))

	id, err := NewStmt(db, "userinfo").SetLogger(log.GetLogger()).Where("id=2").Update(&data)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("affected row:%+v", id)
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

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	mock.ExpectExec("").WillReturnResult(sqlmock.NewResult(1, 0))

	id, err := NewStmt(db, "userinfo").Insert(&data)
	if err != nil {
		t.Fatal(err.Error())
	}

	t.Logf("new id:%+v", id)
}

func TestORMSubStruct(t *testing.T) {
	site := []struct {
		ID     int64 `db_defult:"auto"`
		Name   string
		UserID sql.NullInt64
		List   struct {
			ID   int64
			Name string
		} `db_table:"one"`

		App struct {
			ID   int64
			Name string
		} `db_table:"one" db_relation_field:"application_id"`

		Filter []struct {
			ID   int64
			Key1 string
			Key2 string
		} `db_table:"more"`
	}{}

	sql := "select site.id, site.name, site.user_id, list.id, list.name, app.id, app.name from site,list,app where site.list_id = list.id and site.application_id = app.id"

	str, err := NewStmt(nil, "site").sqlQueryBuilder(&site)
	if err != nil {
		t.Fatal(err.Error())
	}
	if str != sql {
		t.Fatalf("\n\texpect:%s\n\t  recv:%s", sql, str)
	}

	t.Logf("sql:%+v", str)
}
