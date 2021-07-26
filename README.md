## Crab 开发必备库  
![Logo](https://raw.githubusercontent.com/dearcode/crab/master/doc/logo.png)   

[![codecov](https://codecov.io/gh/dearcode/crab/branch/master/graph/badge.svg?token=WKPPEUIHJY)](https://codecov.io/gh/dearcode/crab)
# config  
加载ini格式的配置文件, 支持以;或者#开头的注释
```go
type testConf struct {
	DB struct {
		Domain string
		Port   int `default:"9088"`
		Enable bool
	}
    aaa int
}

var conf testConf
if err := LoadConfig(path, &conf); err != nil {
    t.Fatalf(errors.ErrorStack(err))
}
t.Logf("conf:%+v", conf)

```
配置文件  
```ini
 [db]
domain    =jd.com
enable=true
# test comments
;port=3306
```
只要传入对应的ini文件全路径及struct指针就可以了，简单高效.  
运行结果:  
```go
conf:{DB:{Domain:jd.com Port:9088 Enable:true} aaa:0}
```
`注意`:只会解析有访问权限的变量（大写）  


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
//以包名为路径
handler.Server.AddInterface(&user{}, "")
//指定path
handler.Server.AddInterface(&user{}, "/api/user/")

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



