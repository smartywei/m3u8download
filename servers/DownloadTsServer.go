package servers

import (
	"M3u8Tool/drives"
	"fmt"
	"time"
	"encoding/json"
	"strconv"
	"M3u8Tool/config"
	"io"
	"M3u8Tool/core"
	"M3u8Tool/tools"
	"os"
	"path"
	"net/http"
	"crypto/tls"
)

func StartDownloadServer() {

	downloadNum := config.CFG.Section("download").Key("DownloadNum").MustInt(10)

	for i := 0; i < downloadNum; i++ {
		go doDownloadServer()
	}

	fmt.Println(strconv.Itoa(downloadNum) + "个下载进程已启动！")
}

func doDownloadServer() {

	downloadTimeSpace := config.CFG.Section("download").Key("DownloadTimeSpace").MustInt64(3000)

	ticker := time.NewTicker(time.Millisecond * time.Duration(downloadTimeSpace))

	for {
		select {
		case <-ticker.C:
			res, err := drives.RedisRPop("TS_URLS")

			if err != nil {
				continue
			}

			info := map[string]string{}
			json.Unmarshal([]byte(res), &info)
			doDownload(info)
		}
	}
}

func doDownload(downloadInfo map[string]string) {
	//判断此任务是否失败，失败直接跳过
	statusInfo, err := drives.RedisGetKeyValue("JOB_STATUS_" + downloadInfo["id"])
	if err != nil {
		return
	}

	info := map[string]string{}
	err = json.Unmarshal([]byte(statusInfo), &info)
	if err != nil {
		return
	}
	if info["status"] == "fail" {
		return
	}


	downloadFilePath := config.CFG.Section("download").Key("DownloadFilePath").MustString("./file/download")

	// 创建下载目录
	if ok, err := tools.PathExists(downloadFilePath + downloadInfo["id"]); !(ok && err == nil) {
		err := tools.CreateDir(downloadFilePath + downloadInfo["id"])
		if err != nil {
			core.SetJobFail("JOB_STATUS_"+downloadInfo["id"], "4:"+err.Error())
			return
		}
	}


	//开始下载文件
	index, _ := strconv.Atoi(downloadInfo["index"])
	name := fmt.Sprintf("%06d", index)

	f, err := os.Create(downloadFilePath + downloadInfo["id"] + "/" + name + path.Ext(downloadInfo["url"]))
	defer f.Close()

	if err != nil {
		core.SetJobFail("JOB_STATUS_"+downloadInfo["id"], "5:"+err.Error())
		return
	}

	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	response, err := client.Get(downloadInfo["url"])

	if err != nil {
		//重新加入队列，累加一次错误
		core.AccumulateOneFailForTsList(downloadInfo, "1:"+err.Error())
		return
	}

	defer response.Body.Close()

	if response.StatusCode != 200 {
		//重新加入队列，累加一次错误
		core.AccumulateOneFailForTsList(downloadInfo, "2:"+err.Error())
		return
	}

	_, err = io.Copy(f, response.Body)

	if err != nil {
		//重新加入队列，累加一次错误
		core.AccumulateOneFailForTsList(downloadInfo, "3:"+err.Error())
		return
	}

	conn := drives.GetConn()
	num, err := conn.Incr("DownloadTsNum_" + downloadInfo["id"]).Result()
	if err != nil {
		fmt.Println("redis 原子计数错误，id:" + downloadInfo["id"])
		core.AccumulateOneFailForTsList(downloadInfo, "4:"+err.Error())
		return
	}

	if strconv.FormatInt(num, 10) == info["num"] {
		// 加入转换队列
		transformKey := "TRANSFORM_LIST"

		transformFilePath := config.CFG.Section("download").Key("TransformationPath").MustString("./file/transformation/")

		transformInfo, _ := json.Marshal(map[string]string{"id": downloadInfo["id"], "source": downloadFilePath + downloadInfo["id"], "target": transformFilePath, "fail": "0"})

		err := conn.LPush(transformKey, transformInfo).Err()

		if err != nil {
			core.SetJobFail("JOB_STATUS_"+downloadInfo["id"], "6:"+err.Error())
			return
		}
	}
}
