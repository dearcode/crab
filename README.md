# webgo
自己用的web框架, 不支持RESTFul  

# handler  
简单高效的HTTP路由  

# orm  
只支持mysql
``
result := struct {
		ID       int64
		User     string
		Password string
	}{}

	if err = NewStmt(db, "userinfo").Where("id=2").Query(&result); err != nil {
		t.Fatal(err.Error())
	}

  ``

# validation  
这是从beego复制过来的，改了些东西，这个要重新整理下  

