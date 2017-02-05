package orm

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"unicode"

	"github.com/davygeek/log"
	"github.com/juju/errors"
)

var (
	ErrNotFound = errors.New("not found")
)

//Stmt db stmt.
type Stmt struct {
	table  string
	where  string
	sort   string
	order  string
	offset int
	limit  int
	db     *sql.DB
}

//NewStmt new db stmt.
func NewStmt(db *sql.DB, table string) *Stmt {
	return &Stmt{
		table: table,
		db:    db,
	}
}

//Where 添加查询条件
func (s *Stmt) Where(f string, args ...interface{}) *Stmt {
	if len(args) > 0 {
		s.where = fmt.Sprintf(f, args...)
	} else {
		s.where = f
	}
	return s
}

//Sort 添加sort
func (s *Stmt) Sort(sort string) *Stmt {
	s.sort = sort
	return s
}

//Order 添加order
func (s *Stmt) Order(order string) *Stmt {
	s.order = order
	return s
}

//Offset 添加offset
func (s *Stmt) Offset(offset int) *Stmt {
	s.offset = offset
	return s
}

//Limit 添加limit
func (s *Stmt) Limit(limit int) *Stmt {
	s.limit = limit
	return s
}

//SqlQueryBuilder build sql query.
func (s *Stmt) SqlQueryBuilder(result interface{}) (string, error) {
	rt := reflect.TypeOf(result)
	if rt.Kind() != reflect.Ptr {
		return "", fmt.Errorf("result type must be ptr, recv:%v", rt.Kind())
	}

	//ptr
	rt = rt.Elem()
	if rt.Kind() == reflect.Slice {
		rt = rt.Elem()
	} else {
		//只查一条加上limit 1
		s.limit = 1
	}

	//empty struct
	if rt.NumField() == 0 {
		return "", fmt.Errorf("result not found field")
	}

	return s.SqlQuery(rt), nil
}

// SqlQuery 根据条件及结构生成查询sql
func (s *Stmt) SqlQuery(elem reflect.Type) string {
	firstTable := strings.Split(s.table, ",")[0]

	buf := bytes.NewBufferString("select ")

	for i := 0; i < elem.NumField(); i++ {
		name := elem.Field(i).Tag.Get("db")
		if name == "" {
			name = FieldEscape(elem.Field(i).Name)
		}
		if !strings.Contains(name, ".") {
			buf.WriteString(firstTable)
			buf.WriteString(".")
		}
		buf.WriteString(name)
		buf.WriteString(", ")
	}

	buf.Truncate(buf.Len() - 2)
	buf.WriteString(" from ")
	buf.WriteString(s.table)

	if s.where != "" {
		buf.WriteString(" where ")
		buf.WriteString(s.where)
	}

	if s.sort != "" {
		buf.WriteString(" order by ")
		buf.WriteString(s.sort)
		if s.order != "" {
			buf.WriteString(" ")
			buf.WriteString(s.order)
		}
	}

	if s.limit > 0 {
		buf.WriteString(" limit ")
		if s.offset > 0 {
			buf.WriteString(fmt.Sprintf("%d,", s.offset))
		}
		buf.WriteString(fmt.Sprintf("%d", s.limit))
	}

	sql := buf.String()
	log.Debugf("sql:%v", sql)
	return sql
}

// Query 根据传入的result结构，生成查询sql，并返回执行结果， result 必需是一个指向切片的指针.
func (s *Stmt) Query(result interface{}) error {
	rt := reflect.TypeOf(result)

	if rt.Kind() != reflect.Ptr {
		return fmt.Errorf("result type must be ptr, recv:%v", rt.Kind())
	}

	//ptr
	rt = rt.Elem()
	if rt.Kind() == reflect.Slice {
		rt = rt.Elem()
	} else {
		//只查一条加上limit 1
		s.limit = 1
	}

	//empty struct
	if rt.NumField() == 0 {
		return fmt.Errorf("result not found field")
	}

	sql := s.SqlQuery(rt)

	rows, err := s.db.Query(sql)
	if err != nil {
		return errors.Trace(err)
	}
	defer rows.Close()

	rv := reflect.ValueOf(result).Elem()

	for rows.Next() {
		var refs []interface{}
		obj := reflect.New(rt)

		for i := 0; i < obj.Elem().NumField(); i++ {
			refs = append(refs, obj.Elem().Field(i).Addr().Interface())
		}

		if err = rows.Scan(refs...); err != nil {
			return errors.Trace(err)
		}

		if rv.Kind() == reflect.Struct {
			reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj.Elem().Interface()))
			log.Debugf("result %v", result)
			return nil
		}

		rv = reflect.Append(rv, obj.Elem())
	}

	if rv.Kind() == reflect.Struct || rv.Len() == 0 {
		return errors.Trace(ErrNotFound)
	}

	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(rv.Interface()))
	log.Debugf("result %v", result)

	return nil
}

//Add 添加数据
func Add(db *sql.DB, table string, data interface{}) (int64, error) {
	rt := reflect.TypeOf(data)
	rv := reflect.ValueOf(data)

	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	if rt.NumField() == 0 {
		return 0, fmt.Errorf("data not found field")
	}

	bs := bytes.NewBufferString("insert into ")
	bs.WriteString(table)
	bs.WriteString(" (")

	ds := bytes.NewBufferString(") values (")

	var refs []interface{}

	for i := 0; i < rt.NumField(); i++ {
		if rt.Field(i).PkgPath != "" && !rt.Field(i).Anonymous { // unexported
			continue
		}
		def := rt.Field(i).Tag.Get("db_default")
		if def == "auto" {
			continue
		}
		name := rt.Field(i).Tag.Get("db")
		if name == "" {
			name = FieldEscape(rt.Field(i).Name)
		}

		bs.WriteString(name)
		bs.WriteString(", ")

		if def != "" {
			ds.WriteString(def)
			ds.WriteString(", ")
			continue
		}

		ds.WriteString("?, ")
		refs = append(refs, rv.Field(i).Interface())
	}
	bs.Truncate(bs.Len() - 2)
	ds.Truncate(ds.Len() - 2)
	bs.WriteString(ds.String())
	bs.WriteString(") ")
	sql := bs.String()
	log.Debugf("sql:%v", sql)
	r, err := db.Exec(sql, refs...)
	if err != nil {
		return 0, errors.Trace(err)
	}
	return r.LastInsertId()
}

//FieldEscape 转换为小写下划线分隔
func FieldEscape(k string) string {
	buf := []byte{}
	up := true
	for _, c := range k {
		if unicode.IsUpper(c) {
			if !up {
				buf = append(buf, '_')
			}
			c += 32
			up = true
		} else {
			up = false
		}

		buf = append(buf, byte(c))
	}
	return string(buf)
}
