package bbb4go

//-----------------------------------------------------------------------------
// 建立会议室返回XML的数据结构, 即create接口调用的返回值实例
type createMeetingResponse struct {
	Returncode string  `xml:"returncode"` // 是否成功
	Meeting    meeting `xml:"meeting"`
}

type meeting struct {
	MeetingID            string `xml:"meetingID"`            // 会议ID
	CreateTime           string `xml:"createTime"`           // 会议创建时间
	AttendeePW           string `xml:"attendeePW"`           // 与会者密码
	ModeratorPW          string `xml:"moderatorPW"`          // 会议管理员密码
	HasBeenForciblyEnded string `xml:"hasBeenForciblyEnded"` // 是否可以被强制结束
	MessageKey           string `xml:"messageKey"`           // 返回消息Key
	Message              string `xml:"message"`              // 返回消息
}

//-----------------------------------------------------------------------------
// 检查会议室是否在运行返回XML的数据结构, 即isMeetingRunning接口调用的返回值实例
type isMeetingRunningResponse struct {
	ReturnCode string `xml:"returncode"` // 是否成功
	Running    string `xml:"running"`    // 会议室状态
}
