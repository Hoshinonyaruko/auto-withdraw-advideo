package structs

type AccessToken struct {
	SelfID string `yaml:"self_id"`
	Token  string `yaml:"token"`
}

type Settings struct {
	Port                    string        `yaml:"port"`
	WsPath                  string        `yaml:"wspath"`
	Wstoken                 string        `yaml:"wstoken"`
	HttpPaths               []string      `yaml:"paths"`
	HttpPathsAccessTokens   []AccessToken `yaml:"access_tokens"`
	VideoSecondLimit        int           `yaml:"video_second_limit"`
	CheckVideoQRCode        bool          `yaml:"check_video_qrcode"`
	QRLimit                 int           `yaml:"qr_limit"`
	WithdrawNotice          string        `yaml:"withdraw_notice"`
	OnEnableVideoCheck      string        `yaml:"on_enable_video_check"`
	OnDisableVideoCheck     string        `yaml:"on_disable_video_check"`
	OnEnablePicCheck        string        `yaml:"on_enable_pic_check"`
	OnDisablePicCheck       string        `yaml:"on_disable_pic_check"`
	SetGroupKick            bool          `yaml:"set_group_kick"`
	KickAndRejectAddRequest bool          `yaml:"kick_and_reject_add_request"`
	WithdrawWords           []string      `yaml:"withdraw_words"`
}

// Message represents a standardized structure for the incoming messages.
type Message struct {
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params"`
	Echo   interface{}            `json:"echo,omitempty"`
}

type MessageEvent struct {
	PostType    string      `json:"post_type"`
	MessageType string      `json:"message_type"`
	Time        int64       `json:"time"`
	SelfID      int64       `json:"self_id"`
	SubType     string      `json:"sub_type"`
	Message     interface{} `json:"message"`
	RawMessage  string      `json:"raw_message"`
	Sender      struct {
		Age      int    `json:"age"`
		Area     string `json:"area"`
		Card     string `json:"card"`
		Level    string `json:"level"`
		Nickname string `json:"nickname"`
		Role     string `json:"role"`
		Sex      string `json:"sex"`
		Title    string `json:"title"`
		UserID   int64  `json:"user_id"`
	} `json:"sender"`
	UserID    int64 `json:"user_id"`
	Anonymous *struct {
	} `json:"anonymous"`
	Font       int   `json:"font"`
	GroupID    int64 `json:"group_id"`
	MessageSeq int64 `json:"message_seq"`
	MessageID  int64 `json:"message_id"`
}

type MetaEvent struct {
	PostType      string `json:"post_type"`
	MetaEventType string `json:"meta_event_type"`
	Time          int64  `json:"time"`
	SelfID        int64  `json:"self_id"`
	Interval      int    `json:"interval"`
	Status        struct {
		AppEnabled     bool  `json:"app_enabled"`
		AppGood        bool  `json:"app_good"`
		AppInitialized bool  `json:"app_initialized"`
		Good           bool  `json:"good"`
		Online         bool  `json:"online"`
		PluginsGood    *bool `json:"plugins_good"`
		Stat           struct {
			PacketReceived  int   `json:"packet_received"`
			PacketSent      int   `json:"packet_sent"`
			PacketLost      int   `json:"packet_lost"`
			MessageReceived int   `json:"message_received"`
			MessageSent     int   `json:"message_sent"`
			DisconnectTimes int   `json:"disconnect_times"`
			LostTimes       int   `json:"lost_times"`
			LastMessageTime int64 `json:"last_message_time"`
		} `json:"stat"`
	} `json:"status"`
}

type NoticeEvent struct {
	GroupID    int64  `json:"group_id"`
	NoticeType string `json:"notice_type"`
	OperatorID int64  `json:"operator_id"`
	PostType   string `json:"post_type"`
	SelfID     int64  `json:"self_id"`
	SubType    string `json:"sub_type"`
	Time       int64  `json:"time"`
	UserID     int64  `json:"user_id"`
}

type RobotStatus struct {
	SelfID          int64  `json:"self_id"`
	Date            string `json:"date"`
	Online          bool   `json:"online"`
	MessageReceived int    `json:"message_received"`
	MessageSent     int    `json:"message_sent"`
	LastMessageTime int64  `json:"last_message_time"`
	InvitesReceived int    `json:"invites_received"`
	KicksReceived   int    `json:"kicks_received"`
	DailyDAU        int    `json:"daily_dau"`
}
