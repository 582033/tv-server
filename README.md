# IPTV 服务器

一个简单的 IPTV 管理平台，支持流媒体验证、智能分类、个性化收藏，让您的 IPTV 资源管理更轻松。

| - | - | - |
|:---:|:---:|:---:|
| ![首页](https://raw.githubusercontent.com/582033/tv-server/refs/heads/main/docs/homepage.jpg) | ![频道页](https://raw.githubusercontent.com/582033/tv-server/refs/heads/main/docs/channel.jpg)| ![详情页](https://raw.githubusercontent.com/582033/tv-server/refs/heads/main/docs/detail.jpg)|


## 功能特点
* 🚀 M3U 流媒体本地验证与代理
* 📱 响应式界面设计
* ⭐ 个性化收藏管理
* 💾 支持 MongoDB/SQLite


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
* [x] 查看数据库中的详细视频信息
* [x] 仿照 `https://pleyr.net/en/play` 播放界面，左边为频道列表，右边为视频列表
* [x] 在返回m3u8链接列表时，同时返回延迟率信息
* [ ] 管理导入的视频分类
* [ ] 可收藏频道列表,以及使用列表生成自己的m3u视频源
* [ ] 收藏管理
* [ ] **接入AI，自动优化频道名称及查找源 **