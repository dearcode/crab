# Petrel
petrel海燕，自己用的web框架, 不支持RESTFul，不是用来写接口的  

# handler  
简单高效的HTTP路由，支持指定接口函数，支持自动注册接口  
指定接口注册示例:  
```go
handler.Server.AddHandler(handler.GET, "/test/", false, onTestGet)
handler.Server.AddHandler(handler.POST, "/test/", false, onTestPost)
handler.Server.AddHandler(handler.DELETE, "/test/", false, onTestDelete)
```
自动注册接口示例：  
```go
handler.Server.AddInterface(&user{})

type user struct {
}

//DoGet 默认get方法
func (u *user) DoGet(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Get user"))
}

//DoPost 默认post方法
func (u *user) DoPost(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Post user"))
}
```   

# orm    
只支持mysql  
查询示例
```go
result := struct {
	ID       int64
	User     string
	Password string
}{}

if err = NewStmt(db, "userinfo").Where("id=2").Query(&result); err != nil {
	t.Fatal(err.Error())
}

```  
修改示例
```go
data := struct {
	User     string
	Password string
}{
	User:     fmt.Sprintf("new_user_%d", time.Now().Unix()),
	Password: fmt.Sprintf("new_password_%d", time.Now().Unix()),
}

id, err := NewStmt(db, "userinfo").Where("id=2").Update(&data)
if err != nil {
	t.Fatal(err.Error())
}

```  
添加示例  
```go
data := struct {
	ID       int64 `db_defult:"auto"`
	User     string
	Password string
}{
	User:     fmt.Sprintf("user_%d", time.Now().Unix()),
	Password: fmt.Sprintf("password_%d", time.Now().Unix()),
}

id, err := NewStmt(db, "userinfo").Insert(&data)
if err != nil {
	t.Fatal(err.Error())
}

```  

# validation  
这是从beego复制过来的，改了些东西，这个要重新整理下  

