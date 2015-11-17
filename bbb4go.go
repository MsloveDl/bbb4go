// bbb4go project bbb4go.go
package bbb4go

import (
	"crypto/sha1"
	"encoding/xml"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
)

/*
* 建立会议室需要的参数
 */
type CreateMeetingParam struct {
	Name_                   string // 必填, 会议名称;
	MeetingID_              string // 必填, 会议ID, 必须是唯一的;
	AttendeePW_             string // 必填, 与会者密码;
	ModeratorPW_            string // 必填, 会议管理员密码;
	Welcome                 string // 可选, 欢迎语, 具有格式化功能, 参考说明;
	DialNumber              string // 可选, 可通过电话直接拨入语音会议的号码;
	VoiceBridge             string // 可选, 通过电话拨入语音会议时需要输入的PIN码;
	WebVoice                string // 可选, 通过Web方式加入语音会议时需要输入的PIN码;
	LogoutURL               string // 可选, 退出会议后的URL;
	Record                  string // 可选, 是否录制会议, 默认为false;
	Duration                int    // 可选, 会议时长(分钟), 超过时间后会议自动结束. 默认为0;
	Meta                    string // 可选, 会议的元信息描述;
	ModeratorOnlyMessage    string // 可选, 显示一个消息给所有公共聊天室的人;
	AutoStartRecording      bool   // 可选, 当第一个用户进入时自动开始录制会议, 默认为false;
	AllowStartStopRecording bool   // 可选, 是否允许用户启动/停止录制, 默认为true;
}

/*
* 加入会议室需要的参数
 */
type JoinMeetingParam struct {
	FullName_    string // 必填, 用户名
	MeetingID_   string // 必填, 试图加入的会议ID
	Password_    string // 必填, 会议室密码, 这里特指与会者密码, 如果传入管理员密码则以管理员身份进入
	CreateTime   string // 可选, 会议室创建时间, 用来匹配MeetingID, 避免同一个参会者多次进入
	UserID       string // 可选, 标识用户身份的ID, 在调用GetMeetingInfo时将被返回
	WebVoiceConf string // 可选, VOIP协议扩展
	ConfigToken  string // 可选, 由SetConfigXML调用返回的Token
	AvatarURL    string // 可选, 用户头像URL, 当config.xml中displayAvatar为true时提供
	Redirect     string // 可选, 实验, 当HTML5不可用时, 用来重定向到Flash客户端
	ClientURL    string // 可选, 试验, 用来显示自动以的客户端名称
}

func CreateMeeting(param CreateMeetingParam) string {
	strBaseUrl := "http://10.10.1.217" + "api/create?"

	// 检查必填参数
	if "" == param.Name_ || "" = param.MeetingID_ || "" == param.AttendeePW_; "" == param.ModeratorPW {
		return "PARAM ERROR"
	}

	// 构造必填参数
	name := "name=" + param.Name_                       // 会议名称
	meetingID := "&meetingID=" + param.MeetingID_       // 会议ID
	attendeePW := "&attendeePW=" + param.AttendeePW_    // 与会者密码
	moderatorPW := "&moderatorPW=" + param.ModeratorPW_ // 管理员密码

	var welcome string                 // 欢迎语
	var logoutURL string               // 退出后地址
	var record string                  // 是否可以录制
	var duration string                // 会议时长
	var moderatorOnlyMessage string    // 问候语
	var allowStartStopRecording string // 是否允许启动/停止录制
	var voiceBridge string             // 通过Web加入语音会议时的PIN码

	if "" != param.Welcome {
		welcome = "&welcome=" + param.Welcome
	}

	if "" != param.LogoutURL {
		logoutURL = "&logoutURL=" + param.LogoutURL
	}

	if "" != param.Record {
		record = "&record=" + param.Record
	}

	if nil != param.Duration {
		duration = "&duration=" + strconv.Itoa(param.Duration)
	} else {
		duration = "&duration=" + "0"
	}

	if nil != param.AllowStartStopRecording {
		allowStartStopRecording = "&allowStartStopRecording=" + strconv.
			FormatBool(param.AllowStartStopRecording)
	} else {
		allowStartStopRecording = "&allowStartStopRecording=" + "false"
	}

	if "" != param.ModeratorOnlyMessage {
		moderatorOnlyMessage = "&moderatorOnlyMessage=" + param.ModeratorOnlyMessage
	} else {
		moderatorOnlyMessage = "&moderatorOnlyMessage=" + "我是[" + param.Name_ +
			"]大家好."
	}

	if "" != param.VoiceBridge {
		voiceBridge = "&voiceBridge=" + param.VoiceBridge
	} else {
		// 如果VoiceBridge参数为空, 那么我们分配一个随机数给它
		rand.Seed(9999)
		nTemp := 70000 + rand.Intn(9999)
		voiceBridge = "&voiceBridge=" + strconv.Itoa(nTemp)
	}

	// 合成请求的参数
	createParam := name + meetingID + attendeePW + moderatorPW + welcome +
		voiceBridge + logoutURL + record + duration + moderatorOnlyMessage +
		allowStartStopRecording

	// 2015/11/16
	// 继续合成携带checksum参数的请求
}

/*
* 执行HTTP GET请求, 返回请求结果
* 参数: url, 携带参数的请求地址
* 返回: 请求结果, 如果返回ERROR说明请求过程中出错, 详细信息可以查看log
**/
func HttpGet(url string) string {
	response, err := http.Get(url)

	if nil != err {
		log.Println("HTTP GET ERROR: " + err.Error())
		return "ERROR"
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if nil != err {
		log.Println("HTTP GET ERROR: " + err.Error())
		return "ERROR"
	}

	return string(body)
}

/*
* 将Struct转换为Map格式
* type demo struct {              key    value
*     id string        ----\      id     001
*     name string      ----/      name   名字
* }
* 参数: obj, 需要转换的结构体实例
* 返回: Map类型的结果
**/
func Struct2Map(obj interface{}) map[string]interface{} {
	t := reflect.TypeOF(obj)
	v := reflect.ValueOf(obj)

	var data = make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		data[t.Filed(i).Name] = v.Field(i).Interface()
	}

	return data
}
