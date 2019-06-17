package controllers

import (
	"net/http"
	"fmt"
	"M3u8Tool/core"
	"M3u8Tool/tools"
	"encoding/json"
)

func AddJobController(w http.ResponseWriter, r *http.Request) {

	//验证IP是否在允许的范围内
	if !tools.IpCanAccess(r) {
		w.WriteHeader(403)
		return
	}

	r.ParseForm() //解析参数

	m3u8Url := r.Form["m3u8_url"]
	if len(m3u8Url) <= 0 || len(m3u8Url[0]) <= 0 {
		w.WriteHeader(422)
		fmt.Fprintln(w, "m3u8 url can't is null")
		return
	}

	arr, err := core.GetM3u8HrefList(m3u8Url[0])

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "Parse m3u8 file fail")
		return
	}

	jobKey, err := core.SetTsUrlToRedis(arr)

	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintln(w, "Set ts list to redis fail")
		return
	}

	resJson, _ := json.Marshal(map[string]string{"key": jobKey})
	tools.ReturnJosn(resJson, w, 200)
}
