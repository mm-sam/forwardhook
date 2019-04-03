package main

import "testing"

func Test_SaveConfig(t *testing.T) {

	err, conf := NewConfig(getLocalConfigPath("./config.json"))
	if err != nil {
		t.Log(err)
		t.Fail()
		return
	}
	conf.Listen = "0.0.0.0:8899"
	conf.MaxRetries = 10
	conf.Mappings = []Mapping{
		Mapping{
			Path: "/hello",
			Sites: []string{
				"http://www.baidu.com",
			},
		},
		Mapping{
			Path: "/",
			Sites: []string{
				"http://www.baidu.com",
			},
		},
	}

	err = conf.Save()
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	t.Log("PASS")
}
