package drives

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
)

func ConcatTsToMp4(path string, target string) (*exec.Cmd, error) {
	fmt.Println(1)
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, err
	}

	// 生成拼接ts的txt文件

	txtFile, err := os.Create(path+"/concat.txt")
	defer txtFile.Close()
	if err != nil {
		return nil, nil
	}


	for _, f := range files {
		if f.Name() == "concat.txt"{
			continue
		}
		_,err = txtFile.WriteString("file '"+f.Name()+"'" + "\n")
		if err != nil{
			return nil,nil
		}
	}

	return exec.Command("ffmpeg", "-f","concat","-i", path+"/concat.txt", "-y", target), nil
}
