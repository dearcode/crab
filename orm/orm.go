package orm

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/juju/errors"

	"dearcode.net/crab/log"
	"dearcode.net/crab/meta"
	"dearcode.net/crab/util/str"
)

//Stmt db stmt.
type Stmt struct {
	table  string
	where  string
	sort   string
	order  string
	group  string
	offset int
	limit  int
	raw    string
	db     *sql.DB
	logger *log.Logger
}

//IsNotFound error为not found.
func IsNotFound(err error) bool {
	return errors.Cause(err) == meta.ErrNotFound
}

//NewStmt new db stmt.
func NewStmt(db *sql.DB, table string) *Stmt {
	return &Stmt{
		table: table,
		db:    db,
	}
}

//SetLogger 输出log.
func (s *Stmt) SetLogger(l *log.Logger) *Stmt {
	s.logger = l
	return s
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

//Group 添加group by.
func (s *Stmt) Group(group string) *Stmt {
	s.group = group
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

//sqlQueryBuilder build sql query.
func (s *Stmt) sqlQueryBuilder(result interface{}) (string, error) {
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

	return s.sqlQuery(rt), nil
}

func (s *Stmt) addWhere(w string) {
	if s.where != "" {
		s.where += " and "
	}
	s.where += w
}

//sqlOption where, order, limit
func (s *Stmt) sqlOption(bs *bytes.Buffer) *bytes.Buffer {
	if s.where != "" {
		fmt.Fprintf(bs, " where %s", s.where)
	}

	if s.sort != "" {
		fmt.Fprintf(bs, " order by %s", s.sort)
		if s.order != "" {
			fmt.Fprintf(bs, " %s", s.order)
		}
	}

	if s.group != "" {
		fmt.Fprintf(bs, " group by %s", s.group)
	}

	if s.limit > 0 {
		bs.WriteString(" limit ")
		if s.offset > 0 {
			fmt.Fprintf(bs, "%d,", s.offset)
		}
		fmt.Fprintf(bs, "%d", s.limit)
	}
	return bs
}

// sqlCount 根据条件及结构生成查询sql
func (s *Stmt) sqlCount() string {
	bs := bytes.NewBufferString("select count(*) from ")
	bs.WriteString(s.table)

	s.sqlOption(bs)

	sql := bs.String()
	s.logger.Debugf("sql:%v", sql)
	return sql
}

//sqlColumn 生成查询需要的列，目前只是内部用.
func (s *Stmt) sqlColumn(rt reflect.Type, table string) string {
	bs := bytes.NewBufferString("")

	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		if f.PkgPath != "" && !f.Anonymous { // unexported
			continue
		}
		switch f.Type.Kind() {
		case reflect.Struct:
			if f.Tag.Get("db_table") == "one" {
				s.table += ","
				field := str.FieldEscape(f.Name)
				s.table += field
				bs.WriteString(s.sqlColumn(f.Type, str.FieldEscape(f.Name)))

				if rf := f.Tag.Get("db_relation_field"); rf != "" {
					s.addWhere(fmt.Sprintf("%s.%s = %s.id", table, rf, field))
				} else {
					s.addWhere(fmt.Sprintf("%s.%s_id = %s.id", table, field, field))
				}

				continue
			}
		case reflect.Slice:
			continue
		}

		attr := ""
		name := f.Tag.Get("db")

		if kv := strings.Split(name, ","); len(kv) == 2 {
			name = kv[0]
			attr = kv[1]
		}

		if name == "" {
			name = str.FieldEscape(f.Name)
		}

		switch attr {
		case "distinct":
			bs.WriteString("distinct ")
		}

		if !strings.Contains(name, ".") && !strings.Contains(name, "(") {
			fmt.Fprintf(bs, "%s.", table)
		}
		fmt.Fprintf(bs, "%s, ", name)
	}

	return bs.String()
}

// sqlQuery 根据条件及结构生成查询sql
func (s *Stmt) sqlQuery(rt reflect.Type) string {
	if s.raw != "" {
		return s.raw
	}

	bs := bytes.NewBufferString("select ")
	firstTable := strings.Split(s.table, ",")[0]
	bs.WriteString(s.sqlColumn(rt, firstTable))
	bs.Truncate(bs.Len() - 2)
	bs.WriteString(" from ")
	bs.WriteString(s.table)

	s.sqlOption(bs)
	sql := bs.String()

	s.logger.Debugf("sql:%v", sql)

	return sql
}

func (s *Stmt) firstTable() string {
	if idx := strings.Index(s.table, ","); idx > -1 {
		return s.table[:idx]
	}
	return s.table
}

// addRelation 添加多表关联条件
func (s *Stmt) addRelation(t1, t2 string, id interface{}) *Stmt {
	t1 = str.FieldEscape(t1)
	t2 = str.FieldEscape(t2)
	s.addWhere(fmt.Sprintf("id in (select %s_id from %s_%s_relation where %s_id=%d)", t1, t2, t1, t2, id))
	return s
}

// addOne2More 添加一对多关联条件
func (s *Stmt) addOne2More(t1, t2 string, id interface{}) *Stmt {
	t1 = str.FieldEscape(t1)
	t2 = str.FieldEscape(t2)
	s.addWhere(fmt.Sprintf("%s.%s_id=%d", t1, t2, id))
	return s
}

// Query 根据传入的result结构，生成查询sql，并返回执行结果， result 必需是一个指向切片的指针.
func (s *Stmt) Query(result interface{}) error {
	if result == nil {
		return meta.ErrArgIsNil
	}

	rt := reflect.TypeOf(result)

	if rt.Kind() != reflect.Ptr {
		return errors.Trace(meta.ErrArgNotPtr)
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

	sql := s.sqlQuery(rt)

	rows, err := s.db.Query(sql)
	if err != nil {
		return errors.Annotatef(err, sql)
	}
	defer rows.Close()

	rv := reflect.ValueOf(result).Elem()

	for rows.Next() {
		var refs []interface{}
		obj := reflect.New(rt).Elem()
		var idx int

		for i := 0; i < obj.NumField(); i++ {
			f := rt.Field(i)
			if f.PkgPath != "" && !f.Anonymous { // unexported
				continue
			}

			if f.Name == "ID" {
				idx = len(refs)
			}

			switch f.Type.Kind() {
			case reflect.Struct:
				if f.Tag.Get("db_table") == "one" {
					//一对一，这里代码重复是为了减少交互.
					for j := 0; j < obj.Field(i).NumField(); j++ {
						sf := rt.Field(i).Type.Field(j)
						if sf.PkgPath != "" && !sf.Anonymous { // unexported
							continue
						}
						if sf.Type.Kind() == reflect.Slice {
							continue
						}

						refs = append(refs, obj.Field(i).Field(j).Addr().Interface())
					}
					continue
				}
			case reflect.Slice:
				continue
			}

			refs = append(refs, obj.Field(i).Addr().Interface())
		}

		if err = rows.Scan(refs...); err != nil {
			return errors.Trace(err)
		}

		//一对多
		for i := 0; i < obj.NumField(); i++ {
			f := rt.Field(i)
			if f.PkgPath != "" && !f.Anonymous { // unexported
				continue
			}
			if f.Type.Kind() != reflect.Slice {
				continue
			}

			lr := obj.Field(i).Addr().Interface()
			id := reflect.ValueOf(refs[idx]).Elem().Interface()

			switch f.Tag.Get("db_table") {
			case "more":
				//填充一对多结果，每次去查询
				if err = NewStmt(s.db, str.FieldEscape(f.Name)).addRelation(f.Name, s.firstTable(), id).Query(lr); err != nil {
					if errors.Cause(err) != meta.ErrNotFound {
						return errors.Trace(err)
					}
				}
			case "one2more":
				//填充一对多结果，每次去查询
				if err = NewStmt(s.db, str.FieldEscape(f.Name)).addOne2More(f.Name, s.firstTable(), id).Query(lr); err != nil {
					if errors.Cause(err) != meta.ErrNotFound {
						return errors.Trace(err)
					}
				}
			}
		}

		if rv.Kind() == reflect.Struct {
			reflect.ValueOf(result).Elem().Set(reflect.ValueOf(obj.Interface()))
			s.logger.Debugf("result %#v", result)
			return nil
		}

		rv = reflect.Append(rv, obj)
	}

	if rv.Kind() == reflect.Struct || rv.Len() == 0 {
		return errors.Trace(meta.ErrNotFound)
	}

	reflect.ValueOf(result).Elem().Set(reflect.ValueOf(rv.Interface()))
	s.logger.Debugf("result %v", result)

	return nil
}

//Count 查询总数.
func (s *Stmt) Count() (int64, error) {
	rows, err := s.db.Query(s.sqlCount())
	if err != nil {
		return 0, errors.Trace(err)
	}
	defer rows.Close()

	rows.Next()

	var n int64
	if err = rows.Scan(&n); err != nil {
		return 0, errors.Trace(err)
	}

	return n, nil
}

//sqlInsert 添加数据
func (s *Stmt) sqlInsert(rt reflect.Type, rv reflect.Value) (sql string, refs []interface{}) {
	bs := bytes.NewBufferString("insert into ")
	bs.WriteString(s.table)
	bs.WriteString(" (")

	dbs := bytes.NewBufferString(") values (")

	for i := 0; i < rt.NumField(); i++ {
		key, val, ok := s.sqlStructField(rt.Field(i), false)
		if !ok {
			continue
		}

		bs.WriteString(key)
		bs.WriteString(", ")

		if val != "" {
			dbs.WriteString(val)
			dbs.WriteString(", ")
			continue
		}

		dbs.WriteString("?, ")
		refs = append(refs, rv.Field(i).Interface())
	}

	bs.Truncate(bs.Len() - 2)
	dbs.Truncate(dbs.Len() - 2)

	bs.WriteString(dbs.String())

	bs.WriteString(") ")
	sql = bs.String()
	s.logger.Debugf("sql:%v", sql)
	return
}

//sqlStructField 解析结构体中的字段，根据default及db值内容生成对应key ,value.
func (s *Stmt) sqlStructField(f reflect.StructField, isUpdate bool) (key string, val string, ok bool) {
	if f.PkgPath != "" && !f.Anonymous { // unexported
		return
	}

	switch f.Type.Kind() {
	case reflect.Struct, reflect.Slice, reflect.Ptr:
		return
	}

	if _, ignore := f.Tag.Lookup("db_auto"); ignore {
		return
	}

	if _, ignore := f.Tag.Lookup("db_ignore"); ignore {
		return
	}

	if key, ok = f.Tag.Lookup("db"); !ok {
		key = str.FieldEscape(f.Name)
	}

	//insert 解析default
	if !isUpdate {
		if val, ok = f.Tag.Lookup("db_default"); ok {
			return
		}
	}

	if val, ok = f.Tag.Lookup("db_const"); ok {
		return
	}

	return key, "", true
}

// sqlUpdate 根据条件及结构生成update sql
func (s *Stmt) sqlUpdate(rt reflect.Type, rv reflect.Value) (sql string, refs []interface{}) {
	bs := bytes.NewBufferString(fmt.Sprintf("update `%s` set ", s.table))

	for i := 0; i < rt.NumField(); i++ {
		key, val, ok := s.sqlStructField(rt.Field(i), true)
		if !ok {
			continue
		}

		if val != "" {
			fmt.Fprintf(bs, "`%s`=%s, ", key, val)
			continue
		}

		fmt.Fprintf(bs, "`%s`=?, ", key)

		refs = append(refs, rv.Field(i).Interface())
	}

	bs.Truncate(bs.Len() - 2)

	return s.sqlOption(bs).String(), refs
}

//Update sql update db.
func (s *Stmt) Update(data interface{}) (int64, error) {
	if data == nil {
		return 0, errors.Trace(meta.ErrArgIsNil)
	}
	rt := reflect.TypeOf(data)
	rv := reflect.ValueOf(data)

	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	if rt.NumField() == 0 {
		return 0, errors.Trace(meta.ErrFieldNotFound)
	}

	sql, refs := s.sqlUpdate(rt, rv)
	s.logger.Debugf("sql:%v, vals:%#v", sql, refs)
	r, err := s.db.Exec(sql, refs...)
	if err != nil {
		return 0, errors.Trace(err)
	}
	return r.RowsAffected()
}

//Insert sql update db.
func (s *Stmt) Insert(data interface{}) (int64, error) {
	if data == nil {
		return 0, errors.Trace(meta.ErrArgIsNil)
	}
	rt := reflect.TypeOf(data)
	rv := reflect.ValueOf(data)

	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
		rv = rv.Elem()
	}

	if rt.NumField() == 0 {
		return 0, errors.Trace(meta.ErrFieldNotFound)
	}

	sql, refs := s.sqlInsert(rt, rv)
	r, err := s.db.Exec(sql, refs...)
	if err != nil {
		return 0, errors.Trace(err)
	}

	return r.LastInsertId()
}

// BatchInsert 数组插入
func (s *Stmt) BatchInsert(data []interface{}) ([]int64, error) {
	fmt.Printf("data:%#v\n", data)
	if len(data) == 0 {
		return nil, errors.Trace(meta.ErrArgIsNil)
	}

	txn, err := s.db.Begin()
	if err != nil {
		return nil, errors.Trace(err)
	}

	defer func() {
		if err != nil {
			txn.Rollback()
			return
		}
		txn.Commit()
	}()

	var ir sql.Result
	var ids []int64

	fmt.Printf("data:%#v\n", data)
	for _, d := range data {
		rt := reflect.TypeOf(d)
		rv := reflect.ValueOf(d)

		if rt.Kind() == reflect.Ptr {
			rt = rt.Elem()
			rv = rv.Elem()
		}

		if rt.Kind() != reflect.Struct || rt.NumField() == 0 {
			s.logger.Errorf("kind:%v, type:%v", rt.Kind(), rt)
			return nil, errors.Trace(meta.ErrFieldNotFound)
		}

		stmt, refs := s.sqlInsert(rt, rv)

		if ir, err = txn.Exec(stmt, refs...); err != nil {
			return nil, errors.Trace(err)
		}

		id, _ := ir.LastInsertId()

		ids = append(ids, id)
	}

	return ids, nil
}

//Exec 保留的原始执行接口.
func (s *Stmt) Exec(query string, args ...interface{}) (int64, error) {
	rs, err := s.db.Exec(query, args...)
	if err != nil {
		return -1, errors.Trace(err)
	}
	return rs.RowsAffected()
}

//Raw 原始sql.
func (s *Stmt) Raw(query string, args ...interface{}) *Stmt {
	s.raw = fmt.Sprintf(query, args...)
	return s
}
