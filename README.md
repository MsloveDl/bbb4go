#bbb4go
##简介
bbb4go 是Bigbluebutton在线会议室系统标准接口协议的Go语言实现调用库. 
在封装中, 我们对在线会议室进行了抽象(MeetingRoom), 将在线会议系统以会议室为最小单位
进行管理.

##安装
您需要安装Go 1.3+ 以确保所有功能的正常使用
通过 go get 下载安装库
$ go get github.com/MsloveDl/bbb4go

##使用
在源码中添加引用即可使用

package main

import "github.com/MsloveDl/bbb4go"

func main() {
    var meetingRoom = bbb4go.MeetingRoom{}
}

##其他
github.com/MsloveDl/bbb4go/models package
包含了所有的模板类, 即库的基本数据结构
github.com/MsloveDl/bbb4go/config package
包含了所有配置信息, 如bbb服务器的私钥
