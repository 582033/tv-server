# IPTV 服务器

一个简单的 IPTV M3U 验证和代理服务器，用于验证和优化 IPTV 播放列表。

![示例界面](https://raw.githubusercontent.com/582033/tv-server/refs/heads/main/docs/example.jpg)


## 功能特点
- 使用本地网络验证,保证视频源的稳定性(所以不要放在公网服务器上)
- 支持多个 M3U 播放列表合并
- 验证 M3U 播放列表中的链接有效性
- 指定延迟速率过滤
- 支持 IPv4 和 IPv6 地址


## 运行

### 方式1: 编译运行
* 复制配置文件`config/dev.json`，并修改其中的mongodb配置为你的mongodb配置
* 编译运行
```
go build -o tv-server main.go
./tv-server -c {$configPath} //例如 ./tv-server -c ./config.json
```

### 方式2: 容器运行
* 复制配置文件`config/dev.json`，并修改其中的mongodb配置为你的mongodb配置
* 运行容器
```
docker run -d -p 8080:8080 -v {configPath}:/config.json tv-server
```
** 注意：如果配置文件中的server.port端口项做了修改，容器内部端口也会随之修改，否则会报错。**

### 方式3: docker-compose
* 复制配置文件`config/dev.json`，重命名为`config.json`,与`docker-compose.yaml`文件同级
* 运行容器
```
docker-compose up -d
```
** 注意：此方法运行时，配置文件中的mongodb配置无需修改 **

## todo
* 增加ffprobe验证
* 查看数据库中的详细视频信息
* 视频分类重命名
* 仿照 `https://pleyr.net/en/play` 播放界面，左边为频道列表，右边为视频列表
* 可收藏频道列表生成自己的m3u视频源
** 接入AI，自动优化频道名称及查找源 **
