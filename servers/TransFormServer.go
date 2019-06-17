package servers

import (
	"fmt"
	"time"
	"encoding/json"
	"M3u8Tool/drives"
	"M3u8Tool/tools"
	"M3u8Tool/core"
	"M3u8Tool/config"
	"strconv"
)

func StartTransFormServer() {

	transformationNum := config.CFG.Section("transformation").Key("TransformationNum").MustInt(10)

	for i := 0; i < transformationNum; i++ {
		go doTransFormServer()
	}

	fmt.Println(strconv.Itoa(transformationNum) + "个转换进程已启动！")
}

func doTransFormServer() {

	transformTimeSpace := config.CFG.Section("transformation").Key("TransformTimeSpace").MustInt64(3000)

	ticker := time.NewTicker(time.Millisecond * time.Duration(transformTimeSpace))

	for {
		select {
		case <-ticker.C:
			res, err := drives.RedisRPop("TRANSFORM_LIST")

			if err != nil {
				continue
			}

			info := map[string]string{}
			json.Unmarshal([]byte(res), &info)
			doTransForm(info)
		}
	}
}

func doTransForm(transFormInfo map[string]string) {
	//判断此任务是否失败，失败直接跳过
	statusInfo, err := drives.RedisGetKeyValue("JOB_STATUS_" + transFormInfo["id"])
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

	// 创建转换目录
	if ok, err := tools.PathExists(transFormInfo["target"]); !(ok && err == nil) {
		err := tools.CreateDir(transFormInfo["target"])
		if err != nil {
			core.SetJobFail("JOB_STATUS_"+transFormInfo["id"], "2_1:"+err.Error())
		}
	}

	//开始转换文件
	cmd, err := drives.ConcatTsToMp4(transFormInfo["source"], transFormInfo["target"]+transFormInfo["id"]+".mp4")

	if err != nil {
		core.SetJobFail("JOB_STATUS_"+transFormInfo["id"], "2_2:"+err.Error())
		return
	}

	err = cmd.Run()

	if err != nil {
		core.SetJobFail("JOB_STATUS_"+transFormInfo["id"], "2_4:"+err.Error())
		return
	}

	if !cmd.ProcessState.Success() {
		core.AccumulateOneFailForTransformList(transFormInfo, "2_3:ffmpeg drive error")
	}

}
