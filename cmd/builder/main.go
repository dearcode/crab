package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	//	"reflect"
	"strings"
	"unicode"

	//	"github.com/davygeek/log"
	"github.com/juju/errors"
)

/*
var (
	ErrNotFound = errors.New("not found")
)

type dbQueryArgs struct {
	table  string
	where  string
	sort   string
	order  string
	offset int
	limit  int
}

//QueryOption 查询选项
type QueryOption func(*dbQueryArgs)

func queryTable(table string) QueryOption {
	return func(a *dbQueryArgs) {
		a.table = table
	}
}

//QueryWhere 添加查询条件
func QueryWhere(f string, args ...interface{}) QueryOption {
	return func(a *dbQueryArgs) {
		if len(args) > 0 {
			a.where = fmt.Sprintf(f, args...)
		} else {
			a.where = f
		}
	}
}

//QuerySort 添加sort
func QuerySort(sort string) QueryOption {
	return func(a *dbQueryArgs) {
		a.sort = sort
	}
}

//QueryOrder 添加order
func QueryOrder(order string) QueryOption {
	return func(a *dbQueryArgs) {
		a.order = order
	}
}

//QueryOffset 添加offset
func QueryOffset(offset int) QueryOption {
	return func(a *dbQueryArgs) {
		a.offset = offset
	}
}

//QueryLimit 添加limit
func QueryLimit(limit int) QueryOption {
	return func(a *dbQueryArgs) { a.limit = limit }
}

// SqlQuery 根据条件及结构生成查询sql
func SqlQuery(result interface{}, opts ...QueryOption) (sql string, elem reflect.Type, err error) {
	var args dbQueryArgs
	for _, o := range opts {
		o(&args)
	}

	rt := reflect.TypeOf(result)

	if rt.Kind() != reflect.Ptr {
		err = fmt.Errorf("result type must be ptr, recv:%v", rt.Kind())
		return
	}
	elem = rt.Elem()
	if elem.Kind() == reflect.Slice {
		elem = elem.Elem()
	} else {
		//只查一条加上limit 1
		args.limit = 1
	}
	if elem.NumField() == 0 {
		err = fmt.Errorf("result not found field")
		return
	}

	buf := bytes.NewBufferString("select ")

	firstTable := strings.Split(args.table, ",")[0]

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
	buf.WriteString(args.table)

	if args.where != "" {
		buf.WriteString(" where ")
		buf.WriteString(args.where)
	}

	if args.sort != "" {
		buf.WriteString(" order by ")
		buf.WriteString(args.sort)
		if args.order != "" {
			buf.WriteString(" ")
			buf.WriteString(args.order)
		}
	}

	if args.limit > 0 {
		buf.WriteString(" limit ")
		if args.offset > 0 {
			buf.WriteString(fmt.Sprintf("%d,", args.offset))
		}
		buf.WriteString(fmt.Sprintf("%d", args.limit))
	}

	sql = buf.String()
	log.Debugf("sql:%v", sql)
	return
}

// Query 根据传入的result结构，生成查询sql，并返回执行结果， result 必需是一个指向切片的指针.
func Query(db *sql.DB, table string, result interface{}, opts ...QueryOption) error {
	opts = append(opts, queryTable(table))
	sql, elem, err := SqlQuery(result, opts...)
	if err != nil {
		return errors.Trace(err)
	}

	rows, err := db.Query(sql)
	if err != nil {
		return errors.Trace(err)
	}
	defer rows.Close()

	rv := reflect.ValueOf(result).Elem()

	for rows.Next() {
		var refs []interface{}
		obj := reflect.New(elem)

		for i := 0; i < obj.Elem().NumField(); i++ {
			refs = append(refs, obj.Elem().Field(i).Addr().Interface())
		}

		if err := rows.Scan(refs...); err != nil {
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
*/

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

//TrimSpace 删除头尾的空字符，空格，换行之类的东西, 具体在unicode.IsSpac.
func TrimSpace(raw string) string {
	s := strings.TrimLeftFunc(raw, unicode.IsSpace)
	if s != "" {
		s = strings.TrimRightFunc(s, unicode.IsSpace)
	}
	return s
}

// TrimSplit 按sep拆分，并去掉空字符.
func TrimSplit(raw, sep string) []string {
	var ss []string

	s := TrimSpace(raw)
	if s == "" {
		return ss
	}

	ss = strings.Split(s, sep)
	i := 0
	for _, s := range ss {
		s = TrimSpace(s)
		if s != "" {
			ss[i] = s
			i++
		}
	}

	return ss[:i]
}

type rowInfo struct {
	Field   string
	Type    string
	Null    string
	Key     string
	Default sql.NullString
	Extra   string
}

type tableInfo struct {
	rows []rowInfo
}

func parseTable(db *sql.DB, table string) (*tableInfo, error) {
	sql := fmt.Sprintf("desc `%s`", table)
	rows, err := db.Query(sql)
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer rows.Close()

	t := &tableInfo{}

	for rows.Next() {
		var ri rowInfo
		if err = rows.Scan(&ri.Field, &ri.Type, &ri.Null, &ri.Key, &ri.Default, &ri.Extra); err != nil {
			return nil, errors.Trace(err)
		}
		t.rows = append(t.rows, ri)
	}

	return t, nil
}

func buildStruct(t *tableInfo) (string, error) {
	buf := bytes.NewBufferString("")
	for _, r := range t.rows {
		switch {
		case strings.Contains(r.Type, "int"):
		}
		fmt.Fprintf(buf, "%+v", r)
	}
	return buf.String(), nil
}

func main() {
	ip := flag.String("h", "127.0.0.1", "Connect to host.")
	port := flag.Int("P", 3306, "Port number to use for connection.")
	name := flag.String("D", "", "Database to use.")
	user := flag.String("u", "root", "User for login.")
	pass := flag.String("p", "", "Password fro login.")
	charset := flag.String("c", "utf8", "Set the default character set.")
	table := flag.String("t", "", "Table names.")

	flag.Parse()

	dbc := NewDB(*ip, *port, *name, *user, *pass, *charset, "10")
	db, err := dbc.GetConnection()
	if err != nil {
		panic(errors.ErrorStack(err))
	}

	for _, t := range TrimSplit(*table, ",") {
		t, err := parseTable(db, t)
		if err != nil {
			panic(errors.ErrorStack(err))
		}
		dat, err := buildStruct(t)
		if err != nil {
			panic(errors.ErrorStack(err))
		}
		fmt.Printf("data:%v\n", dat)
	}

}
