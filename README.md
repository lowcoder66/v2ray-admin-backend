# v2ray-admin-backend
本项目是使用 [go](https://github.com/golang/go) 实现的 [v2ray-admin](https://github.com/lowcoder66/v2ray-admin) 后端，
主要使用 [v2ray](https://github.com/v2fly/v2ray-core) 中的 `v2ctl` 命令进行用户的添加、删除，以及流量统计等服务，并存储至数据库

> 这是我第一个go语言项目，如果您发现了一些很蠢的代码，还请谅解😂

## 功能描述
* 授权用户登录
* 用户连接配置模板
* 提供服务端配置链接
* 流量数据仪表
* 用户管理

## 运行之前
1. 修改位于 `./conf` 中的 `conf.toml` 文件，根据实际情况配置服务端口、数据库、缓存、发件箱等
2. 修改位于 `./resource` 中的 `v2ray-server-config.json` 文件，此文件是提供给 `v2ray` 的url配置模板

❗❗❗本项目以研究学习为目的，请勿用于他途❗❗❗