package core

import (
	"net/url"
	"strings"
	"net/http"
	"io/ioutil"
	"path"
	"M3u8Tool/drives"
	"strconv"
	"time"
	"encoding/json"
	"fmt"
	"M3u8Tool/config"
	"crypto/tls"
)

func GetM3u8HrefList(href string) ([]string, error) {

	var urlPre string
	parseUrl, _ := url.Parse(href)

	if strings.HasPrefix(href, "https") {
		urlPre = "https://" + parseUrl.Host
	} else if strings.HasPrefix(href, "http") {
		urlPre = "http://" + parseUrl.Host
	}

	tr := &http.Transport{
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}

	response, err := client.Get(href)

	if err != nil {
		return []string{}, err
	}

	defer response.Body.Close()

	res, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return []string{}, err
	}

	linList := strings.Split(string(res), "\n")

	var hrefArr []string

	for _, v := range linList {

		if strings.HasPrefix(v, "/") && strings.ToLower(path.Ext(v)) == ".ts" {
			hrefArr = append(hrefArr, urlPre+v)
		} else if strings.HasPrefix(v, "/") && strings.ToLower(path.Ext(v)) == ".m3u8" {
			newList, err := GetM3u8HrefList(urlPre + v)
			if err != nil {
				return []string{}, err
			}
			hrefArr = append(hrefArr, newList...)
		} else if !strings.HasPrefix(v, "#") && strings.ToLower(path.Ext(v)) == ".ts" {
			hrefArr = append(hrefArr, v)
		} else if !strings.HasPrefix(v, "#") && strings.ToLower(path.Ext(v)) == ".m3u8" {
			newList, err := GetM3u8HrefList(v)
			if err != nil {
				return []string{}, err
			}
			hrefArr = append(hrefArr, newList...)
		}
	}

	return hrefArr, nil
}

func SetTsUrlToRedis(TsUrls []string) (string, error) {
	coon := drives.GetConn()
	defer coon.Close()

	key := "TS_URLS"
	id := strconv.FormatInt(time.Now().UnixNano(), 10)
	statusKey := "JOB_STATUS_" + id
	statusInfo, _ := json.Marshal(map[string]string{"status": "wait", "num": strconv.Itoa(len(TsUrls))})
	coon.Set(statusKey, statusInfo, 0)

	for k, v := range TsUrls {
		tsUrlInfo, _ := json.Marshal(map[string]string{"id": id, "url": v, "index": strconv.Itoa(k), "fail": "0"})
		err := coon.LPush(key, tsUrlInfo).Err()
		if err != nil {
			statusInfo, _ = json.Marshal(map[string]string{"status": "fail"})
			coon.Set(statusKey, statusInfo, 0)
			return "", err
		}
	}
	return statusKey, nil
}

func SetJobFail(key string, errInfo string) {
	statusInfo, _ := json.Marshal(map[string]string{"status": "fail", "err": errInfo})
	err := drives.RedisSetKeyValue(key, statusInfo, 0)
	if err != nil {
		//todo error 处理
		fmt.Println("redis 异常，程序中断！")
		panic(err)
	}
}

func AccumulateOneFailForTsList(tsUrlInfo map[string]string, errInfo string) {
	fail := tsUrlInfo["fail"]
	num, err := strconv.Atoi(fail)
	if err != nil {
		SetJobFail("JOB_STATUS_"+tsUrlInfo["id"], "1:"+err.Error())
		return
	}
	if num >= 5 {
		SetJobFail("JOB_STATUS_"+tsUrlInfo["id"], "2:"+errInfo)
		return
	} else {
		num += 1
		tsUrlInfo["fail"] = strconv.Itoa(num)
		redisInfo, _ := json.Marshal(tsUrlInfo)
		err := drives.RedisLPush("TS_URLS", redisInfo)
		if err != nil {
			SetJobFail("JOB_STATUS_"+tsUrlInfo["id"], "3:"+err.Error())
			return
		}
	}
}

func AccumulateOneFailForTransformList(transFormInfo map[string]string, errInfo string) {
	fail := transFormInfo["fail"]
	num, err := strconv.Atoi(fail)
	if err != nil {
		SetJobFail("JOB_STATUS_"+transFormInfo["id"], "1:"+err.Error())
		return
	}

	downloadTsErrorNum := config.CFG.Section("download").Key("DownloadTsErrorNum").MustInt(10)

	if num >= downloadTsErrorNum {
		SetJobFail("JOB_STATUS_"+transFormInfo["id"], "2:"+errInfo)
		return
	} else {
		num += 1
		transFormInfo["fail"] = strconv.Itoa(num)
		redisInfo, _ := json.Marshal(transFormInfo)
		err := drives.RedisLPush("TRANSFORM_LIST", redisInfo)
		if err != nil {
			SetJobFail("JOB_STATUS_"+transFormInfo["id"], "3:"+err.Error())
			return
		}
	}
}
