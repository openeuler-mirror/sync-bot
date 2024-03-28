# sync-bot

#### 介绍

一个处理分支之间同步的工具

#### 文档

[设计文档](docs/design.md)

#### 安装教程

本地编译，可用于检查代码是否有编译问题

```shell
cd sync-bot
go build -o /sync-bo
```

如果碰上`golang.org`上的相关包无法下载，可以通过以下方法换成国内的源

```shell
export GO111MODULE=on
export GOPROXY=https://goproxy.cn,direct
# 注意编译go build命令需要和这个设置在同一个终端环境
```

#### 使用说明

#### 测试说明：

1. 将个人fix bug代码提交到test分支；（系统会自动触发构建并拉起服务）
2. 将测试服务配置到个人仓库webhook：https://sync-bot.test.osinfra.cn/hook
3. webhook密码联系infra@openeuler.sh
4. 在个人仓库测试时，需要将 `openeuler-sync-bot` 这个用户邀请到仓库，并赋予仓库开发者权限。邀请通过联系 [Lostwayzxc](https://gitee.com/lostwayzxc)