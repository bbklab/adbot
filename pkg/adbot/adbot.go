package adbot

// AdbHandler represents the adb host (127.0.0.1:5037) handler
type AdbHandler interface {
	ListAdbDevices() ([]string, error)
	NewDevice(serial string) AdbDeviceHandler
}

// AdbDeviceHandler represents one specified adb device handler
type AdbDeviceHandler interface {
	Serial() (string, error)
	Online() bool
	SysInfo() (*AndroidSysInfo, error)
	BatteryInfo() (*AndroidBatteryInfo, error)
	Run(cmd string, args ...string) (string, error) // similar as: adb -s {id} shell
	IsAwake() bool                                  // is screen awake
	AwakenScreen() error
	ScreenCap() ([]byte, error)
	GotoHome() error
	GoBack() error
	Click(x, y int) error
	Swipe(x1, y1, x2, y2 int) error
	SwipeUpUnlock() error
	SwipeDownShowNotify() error
	CurrentTopActivity() (string, error)
	DumpCurrentUI() ([]*AndroidUINode, error)
	FindUINode(nodes []*AndroidUINode, resourceid, resourcetext string) *AndroidUINode
	FindUINodeAndTapMiddleXY(resourceid, resourcetext string) (int, int, error)
	TailSysLogs() (<-chan string, chan struct{})                     // tail -f syslogs (logcat -d; logcat -c)
	WatchSysEvents(keywords []string) (<-chan string, chan struct{}) // watch keywords in syslogs via TailSysLogs
	ListSysNotifies() []*AndroidSysNotify                            // dumpsys notification -> find `tickerTex`
	ClearSysNotifies() error                                         // pull down, find and tap `clear_all` button
	WatchSysNotifies() (<-chan *AndroidSysNotify, chan struct{})     // watch notification

	// Alipay App
	StartAliPay() error
	GotoAlipayTabHome() error                                                                  // 分页: 首页
	GotoAlipayCharging() error                                                                 // 分页: 首页 -> 收钱
	GotoAlipayChargingAmount() error                                                           // 分页: 首页 -> 收钱 -> 设置金额
	AlipayGenerateChargingAmountQrCode(orderID string, fee int) (*AlipayChargingQrCode, error) // 分页: 首页 -> 收钱 -> 设置金额 -> 输入金额和订单号
	GotoAlipayTabProfile() error                                                               // 分页: 我的
	GotoAlipayListOrder() error                                                                // 分页: 我的 -> 账单
	AlipaySearchOrder(orderID string) (*AlipayOrder, error)                                    // 分页: 我的 -> 账单 -> 搜索
}
