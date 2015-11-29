// bbb4go project bbb4go.go
package bbb4go

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/xml"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
)

// 全局变量
const (
	BaseUrl = "http://10.10.1.217/Bigbluebutton/api/"
	Salt    = "39a5303a1540134de8348021143b927e" // 公钥
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

//-----------------------------------------------------------------------------
// 匹配XML的各结构体
type response struct {
	Returncode string  `xml:"returncode"`
	Meeting    meeting `xml:"meeting"`
}

type meeting struct {
	MeetingID            string `xml:"meetingID"`
	CreateTime           string `xml:"createTime"`
	AttendeePW           string `xml:"attendeePW"`
	ModeratorPW          string `xml:"moderatorPW"`
	HasBeenForciblyEnded string `xml:"hasBeenForciblyEnded"`
	MessageKey           string `xml:"messageKey"`
	Message              string `xml:"message"`
}

//-----------------------------------------------------------------------------

/*
* 根据传入的建立会议室结构体包含的内容创建会议室
* 参数: param, 创建会议室的具体条件
* 返回: 创建成功返回会议室ID, 创建失败返回ERROR及失败内容
**/
func CreateMeeting(param CreateMeetingParam) string {
	// 检查必填参数
	if "" == param.Name_ || "" == param.MeetingID_ ||
		"" == param.AttendeePW_ || "" == param.ModeratorPW_ {
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

	//-----------------------------------------------------------------------------
	// 这里可能有问题, 未做字段内容校验, 如果有错着重检查
	duration = "&duration=" + strconv.Itoa(param.Duration)

	allowStartStopRecording = "&allowStartStopRecording=" + strconv.
		FormatBool(param.AllowStartStopRecording)
	//-----------------------------------------------------------------------------

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

	checksum := GetChecksum("create", createParam, Salt)

	// 发出请求
	createResponse := HttpGet(BaseUrl + "create?" + createParam + "&checksum=" +
		checksum)

	if "ERROR" == createResponse {
		return "ERROR"
	}

	// 解析返回的XML结果, 判断是否成功创建会议室
	xmlResponse := response{}
	err := xml.Unmarshal([]byte(createResponse), &xmlResponse)
	if nil != err {
		log.Println("XML PARSE ERROR: " + err.Error())
		return "ERROR"
	}

	if "SUCCESS" == xmlResponse.Returncode {
		return xmlResponse.Meeting.MeetingID
	} else {
		log.Println("CREATE MEETINGROOM FAILD: " + createResponse)
		return "FAILD"
	}

	return "ERROR"
}

/*
* 根据传入的参数获得要加入的会议室的地址, 获取的地址可以直接进入到会议室当中
* 参数: param, 加入会议室的具体条件
* 返回: 加入指定会议室的URL
**/
func GetJoinURL(param JoinMeetingParam) string {
	if "" == param.FullName_ || "" == param.MeetingID_ ||
		"" == param.Password_ {
		return "PARAM ERROR"
	}

	// 构造必填参数
	fullName := "fullName=" + param.FullName_     // 用户名
	meetingID := "&meetingID=" + param.MeetingID_ // 试图加入的会议ID
	password := "&password=" + param.Password_    // 会议室密码, 这里特指与会者密码, 如果传入管理员密码则以管理员身份进入

	var createTime string  // 会议室创建时间, 用来匹配MeetingID, 避免同一个参会者多次进入
	var userID string      // 标识用户身份的ID, 在调用GetMeetingInfo时将返回
	var configToken string // 有SetConfigXML调用返回的Token
	var avatarURL string   // 用户头像的URL, 当config.xml中displayAvatar为true时提供
	var redirect string    // 当HTML5不可用时, 是否重定向到Flash客户端
	var clientURL string   // 重定向URL

	if "" != param.CreateTime {
		createTime = "&createTime=" + param.CreateTime
	}

	if "" != param.UserID {
		userID = "&userID=" + param.UserID
	}

	if "" != param.ConfigToken {
		configToken = "&configToken=" + param.ConfigToken
	}

	if "" != param.AvatarURL {
		avatarURL = "&avatarURL=" + param.AvatarURL
	}

	if "" != param.ClientURL {
		redirect = "&redirect=true"
		clientURL = "&clientURL=" + param.ClientURL
	}

	// 合成请求参数
	joinParam := fullName + meetingID + password + createTime + userID +
		configToken + avatarURL + redirect + clientURL

	checksum := GetChecksum("join", joinParam, Salt)

	return BaseUrl + "join?" + joinParam + "&checksum=" + checksum
}

/*
* 根据请求的接口, 参数以及公钥生成密文
* 参数: method, 请求的接口
*	   param, 请求携带的参数
*      salt, 服务器提供的公钥
* 返回: 加密后的checksum密文
**/
func GetChecksum(method string, param string, salt string) string {
	private := []byte(method + param + salt)
	ciphertext := sha1.Sum(private)

	return hex.EncodeToString(ciphertext[:])
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
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	var data = make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		data[t.Field(i).Name] = v.Field(i).Interface()
	}

	return data
}
