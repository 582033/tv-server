# IPTV 服务器

一个简单的 IPTV M3U 验证和代理服务器，用于验证和优化 IPTV 播放列表。

![示例界面](https://raw.githubusercontent.com/582033/tv-server/refs/heads/main/docs/example.jpg)


## 功能特点

- 支持多个 M3U 播放列表合并
- 验证 M3U 播放列表中的链接有效性
- 自动过滤延迟高于 1000ms 的链接
- 支持 IPv4 和 IPv6 地址
- 缓存管理，自动清理过期缓存
- 现代化的 Web 界面
  - 支持多个链接输入
  - 实时链接验证
  - 进度显示
  - 一键复制结果
  - 响应式设计

### 运行容器
```
docker run -d -p 8080:8080 tv-server
```

## todo
* 增加滑动块支持调整本地延迟值
