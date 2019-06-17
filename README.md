# m3u8download
m3u8多线程下载ts并合并

需要安装redis和ffmepg

运行 go get gopkg.in/ini.v1

编译之后，直接后台运行即可

单任务下载失败会自动重试，具体配置在config/m3u8Tool.ini里
