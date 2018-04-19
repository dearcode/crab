## Config
支持加载ini格式的配置文件  
支持以;或者#开头的注释  

### example
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


