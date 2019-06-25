package adbot

// AdbHandler represents the adb host (127.0.0.1:5037) handler
type AdbHandler interface {
	ListAdbDevices() ([]string, error)
	WatchAdbEvents() (<-chan *AdbEvent, chan struct{})
	NewDevice(serial string) (AdbDeviceHandler, error)
}

// AdbDeviceHandler represents one specified adb device handler
type AdbDeviceHandler interface {
	Serial() (string, error)
	Exists() bool
	Online() bool
	Reboot() error
	Run(cmd string, args ...string) (string, error) // similar as: adb -s {id} shell
	SysInfo() (*AndroidSysInfo, error)
	BatteryInfo() (*AndroidBatteryInfo, error)
	IsAwake() bool // is screen awake
	AwakenScreen() error
	ScreenCap() ([]byte, error)
	GotoHome() error
	GoBack() error
	Click(x, y int) error
	Swipe(x1, y1, x2, y2 int) error
	CurrentTopActivity() (string, error)
	DumpCurrentUI() ([]*AndroidUINode, error)
	FindUINodeAndClick(resourceid, resourcetext string) (int, int, error)
	TailSysLogs() (<-chan string, chan struct{})                     // tail -f syslogs (logcat -d; logcat -c)
	WatchSysEvents(keywords []string) (<-chan string, chan struct{}) // watch keywords in syslogs via TailSysLogs
	ListSysNotifies() []*AndroidSysNotify                            // dumpsys notification -> find `tickerTex`
	ClearSysNotifies() error                                         // pull down, find and tap `clear_all` button
	WatchSysNotifies() (<-chan *AndroidSysNotify, chan struct{})     // watch notification

	// Alipay App
	StartAliPay() error
	AlipaySearchOrder(orderID string) (*AlipayOrder, error) // 分页: 我的 -> 账单 -> 搜索
}
