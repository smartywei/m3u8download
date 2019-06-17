package servers

import (
	"net/http"
	"log"
	"M3u8Tool/controllers"
	"M3u8Tool/config"
	"fmt"
)

func StartHttpServer() {

	http.HandleFunc("/add_job", controllers.AddJobController)

	port := config.CFG.Section("core").Key("StartPort").MustString("6688")

	fmt.Println("准备监听"+port+"端口")

	err := http.ListenAndServe(":"+port, nil) // 设置监听的端口

	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
