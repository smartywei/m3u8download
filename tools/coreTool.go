package tools

import (
	"M3u8Tool/config"
	"strings"
	"net"
	"net/http"
)

func IpCanAccess(r *http.Request) bool {
	allowIPs := config.CFG.Section("access").Key("AccessIP").Strings(",")

	if !InArray("0.0.0.0", allowIPs) {
		var nowIp = string([]byte(r.RemoteAddr)[0:strings.LastIndex(r.RemoteAddr, ":")])
		var ipCanAccess = false

		if nowIp == "127.0.0.1" {
			nowIp = r.Header.Get("X-Real-IP")
			nowIp = net.ParseIP(nowIp).String()
		}

		if InArray(nowIp, allowIPs) {
			ipCanAccess = true
		}

		if !ipCanAccess {
			return false
		}
	}

	return true
}

func ReturnJosn(josn [] byte,w http.ResponseWriter,status int){
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(josn)
}
