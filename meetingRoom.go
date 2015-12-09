package bbb4go

import (
	"encoding/xml"
	"log"
	"math/rand"
	"net/url"
	"strconv"
)

/*******************************************************************************
* 会议室类, 抽象一个会议室模型, 管理整个会议
*******************************************************************************/
type MeetingRoom struct {
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

	CreateMeetingResponse models.createMeetingResponse // 建立会议室返回信息
	MeetingInfo           getMeetingInfoResponse       // 会议室的当前信息
	Participantses        []Participants               // 会议参与者
}

/*******************************************************************************
* 根据会议室的配置创建会议室, 将返回信息存储在CreateMeetingResponse属性中
* 返回: 创建成功返回会议室ID, 创建失败返回ERROR及失败内容
*******************************************************************************/
func (meetingRoom *MeetingRoom) CreateMeeting() string {
	// 检查必填字段
	if "" == meetingRoom.Name_ || "" == meetingRoom.MeetingID_ ||
		"" == meetingRoom.AttendeePW_ || "" == meetingRoom.ModeratorPW_ {
		log.Println("ERROR: PARAM ERROR.")
		return "ERROR: PARAM ERROR."
	}

	// 根据对象字段构造必填参数
	name := "name=" + url.QueryEscape(meetingRoom.Name_)                       // 会议名称
	meetingID := "&meetingID=" + url.QueryEscape(meetingRoom.MeetingID_)       // 会议ID
	attendeePW := "&attendeePW=" + url.QueryEscape(meetingRoom.AttendeePW_)    // 与会者密码
	moderatorPW := "&moderatorPW=" + url.QueryEscape(meetingRoom.ModeratorPW_) // 管理员密码

	var welcome string                 // 欢迎语
	var logoutURL string               // 退出后地址
	var record string                  // 是否可以录制
	var duration string                // 会议时长
	var moderatorOnlyMessage string    // 问候语
	var allowStartStopRecording string // 是否允许启动/停止录制
	var voiceBridge string             // 通过Web加入语音会议时的PIN码

	if "" != meetingRoom.Welcome {
		welcome = "&welcome=" + url.QueryEscape(meetingRoom.Welcome)
	}

	if "" != meetingRoom.LogoutURL {
		logoutURL = "&logoutURL=" + url.QueryEscape(meetingRoom.LogoutURL)
	}

	if "" != meetingRoom.Record {
		record = "&record=" + url.QueryEscape(meetingRoom.Record)
	}

	//-----------------------------------------------------------------------------
	// 这里可能有问题, 未做字段内容校验, 如果有错着重检查
	duration = "&duration=" + url.QueryEscape(strconv.Itoa(meetingRoom.Duration))

	allowStartStopRecording = "&allowStartStopRecording=" +
		url.QueryEscape(strconv.FormatBool(meetingRoom.AllowStartStopRecording))
	//-----------------------------------------------------------------------------

	if "" != meetingRoom.ModeratorOnlyMessage {
		moderatorOnlyMessage = "&moderatorOnlyMessage=" +
			url.QueryEscape(meetingRoom.ModeratorOnlyMessage)
	} else {
		moderatorOnlyMessage = "&moderatorOnlyMessage=" +
			url.QueryEscape("我是["+meetingRoom.Name_+"]大家好.")
	}

	if "" != meetingRoom.VoiceBridge {
		voiceBridge = "&voiceBridge=" + url.QueryEscape(meetingRoom.VoiceBridge)
	} else {
		// 如果VoiceBridge参数为空, 那么我们分配一个随机数给它
		rand.Seed(9999)
		nTemp := 70000 + rand.Intn(9999)
		voiceBridge = "&voiceBridge=" + url.QueryEscape(strconv.Itoa(nTemp))
	}

	// 合成请求的参数
	createParam := name + meetingID + attendeePW + moderatorPW + welcome +
		voiceBridge + logoutURL + record + duration + moderatorOnlyMessage +
		allowStartStopRecording

	checksum := GetChecksum("create", createParam, SALT)

	// 发出请求
	createResponse := HttpGet(BASE_URL + "create?" + createParam + "&checksum=" +
		checksum)

	if "ERROR" == createResponse {
		log.Println("ERROR: HTTP ERROR.")
		return "ERROR: HTTP ERROR."
	}

	// 解析返回的XML结果, 判断是否成功创建会议室
	err := xml.Unmarshal([]byte(createResponse),
		&meetingRoom.CreateMeetingResponse)

	if nil != err {
		log.Println("XML PARSE ERROR: " + err.Error())
		return "ERROR: XML PARSE ERROR."
	}

	if "SUCCESS" == meetingRoom.CreateMeetingResponse.Returncode {
		log.Println("SUCCESS CREATE MEETINGROOM. MEETING ID: " +
			meetingRoom.CreateMeetingResponse.MeetingID)
		return meetingRoom.CreateMeetingResponse.MeetingID
	} else {
		log.Println("CREATE MEETINGROOM FAILD: " + createResponse)
		return "FAILD"
	}

	return "ERROR: UNKONW."
}

/*******************************************************************************
* 检查当前会议室是否正常运行(开门). 当会议室中没有参与者的时候, 该接口调用将认为此会议室
* 未运行, 返回false
* 返回: true, 会议室运行正常; false, 会议室不存在
*******************************************************************************/
func (meetingRoom *MeetingRoom) IsMeetingRunning() bool {
	if "" == meetingRoom.MeetingID_ {
		log.Println("ERROR: PARAM ERROR.")
		return false
	}

	createParam := "meetingID=" + url.QueryEscape(meetingRoom.MeetingID_)
	checksum := GetChecksum("isMeetingRunning", createParam, SALT)

	createResponse := HttpGet(BASE_URL + "isMeetingRunning?" + createParam +
		"&checksum=" + checksum)

	if "ERROR" == createResponse {
		log.Println("ERROR: HTTP ERROR.")
		return false
	}

	responseXML := isMeetingRunningResponse{}
	err := xml.Unmarshal([]byte(createResponse), &responseXML)

	if nil != err {
		log.Println("XML PARSE ERROR: " + err.Error())
		return false
	}

	if "SUCCESS" == responseXML.ReturnCode {
		log.Println("CALLED SUCCESS.")
		return responseXML.Running
	}

	return false
}

/*******************************************************************************
* 注销会议室, 并且将在会议室中的参与者踢出. 该接口服务器做的是不负责任调用, 我们应该在调
* 用该接口一段时间后进行getMeetingInfo或者isMeetingRunning接口调用, 以对关闭操作进行
* 验证, 确保已经成功执行
* 返回: 调用成功返回true, 否则返回false
*******************************************************************************/
func (meetingRoom *MeetingRoom) End() bool {
	if "" == meetingRoom.MeetingID_ || "" == meetingRoom.ModeratorPW_ {
		log.Println("ERROR: PARAM ERROR.")
		return false
	}

	createParam := "meetingID=" + url.QueryEscape(meetingRoom.MeetingID_) +
		"&password=" + url.QueryEscape(meetingRoom.ModeratorPW_)
	checksum := GetChecksum("end", createParam, SALT)

	createResponse := HttpGet(BASE_URL + "end?" + createParam + "&checksum=" +
		checksum)

	if "ERROR" == createResponse {
		log.Println("ERROR: HTTP ERROR.")
		return false
	}

	responseXML := endResponse{}
	err := xml.Unmarshal([]byte(createResponse), &responseXML)

	if nil != err {
		log.Println("XML PARSE ERROR: " + err.Error())
		return false
	}

	if "SUCCESS" == responseXML.ReturnCode {
		log.Println("END MEETING SUCCESS.")
		return true
	}

	return false
}

/*******************************************************************************
* 获取会议室的详细信息, 并且刷新会议室实体中会议室详细信息的描述
* 返回: 会议室详细信息
*******************************************************************************/
func (meetingRoom *MeetingRoom) GetMeetingInfo() *getMeetingInfoResponse {
	if "" == meetingRoom.MeetingID_ || "" == meetingRoom.ModeratorPW_ {
		log.Println("ERROR: PARAM ERROR.")
		return nil
	}

	createParam := "meetingID=" + url.QueryEscape(meetingRoom.MeetingID_) +
		"&password=" + url.QueryEscape(meetingRoom.ModeratorPW_)
	checksum := GetChecksum("getMeetingInfo", createParam, SALT)

	createResponse := HttpGet(BASE_URL + "getMeetingInfo?" + createParam +
		"&checksum=" + checksum)

	if "ERROR" == createResponse {
		log.Println("ERROR: HTTP ERROR.")
		return nil
	}

	err := xml.Unmarshal([]byte(createResponse), &meetingRoom.MeetingInfo)

	if nil != err {
		log.Println("XML PARSE ERROR: " + err.Error())
		return nil
	}

	if "SUCCESS" == meetingRoom.MeetingInfo.ReturnCode {
		log.Println("GET MEETING INFO SUCCESS.")
		return &meetingRoom.MeetingInfo
	}

	return nil
}
