package main

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const log_time = "2021-"

//以下可以定义自己的关键匹配字
const linuxBroadcast = "[AndroidContolClient.cpp::onBroadcast(L420)] Show App"
const startProc = `Start proc \d*:com.maezia.ezia.weather`
const applicationOnCreate = "sub proc onCreate begin"
const applicationOnCreateEnd = "sub proc onCreate end"
const activityOnCreate = "sub proc main activity onCreate begin"
const activityOnCreateEnd = "sub proc main activity onCreate end"
const appBeginLoading = "sub proc get first weather data begin"
const appShowUI = "sub proc setWeatherDataForView finished"
const appShowImageView = "perAirQualityPixel ="
const appWindowVisible = "onWindowsVisible: com.maezia.ezia.weather"
const lowMemoryKiller = " Killing '"
const appServiceConnected = "sub proc on weather service connected"
const initOkhttp = "sub proc http config..."
const httpRequestTook = "request tookMs ="
const lbsAgentbegin = "[onServiceConnected()]class name = com.maezia.ezia.navi.lbsagent.agent.impl.LBSAgentService"
const lbsAgentEnd = "sub proc begin get navi info end"
const backgroundGC = "Background concurrent copying GC freed"
const queueIdleEvent = "register vska, "

var appColdStartInfo *AppColdStartInfo

var regl = regexp.MustCompile(startProc)
var pidRegl = regexp.MustCompile(`\d{4,5}`)

func getAppColdStart(filePath *string) *list.List {
	return analyseColdStartFile(filePath, 0, 0)
}

type AppColdStartInfo struct {
	LinuxBroadcast         time.Time
	ApplicationOnCreate    time.Time
	ApplicationOnCreateEnd time.Time
	ActivityOnCreate       time.Time
	ActivityOnCreateEnd    time.Time
	AppBeginLoading        time.Time
	AppShowUI              time.Time
	AppShowImageView       time.Time
	AppWindowVisible       time.Time
	AppQueueIdleEvent      time.Time

	AppServiceConnected time.Time

	LBSAgentBegin   *time.Time
	LBSAgentEnd     *time.Time
	InitOkHttpBegin *time.Time
	InitOkHttpEnd   *time.Time

	HttpRequestTookMs string

	LowMemKillList   *list.List
	BackgroundGCList *list.List

	pid int
}

func (coldStartInfo *AppColdStartInfo) String() string {
	return fmt.Sprintf(
		`AppColdStartInfo = 
pid = %d
[
Linux_broadcast     =%s, 
ApplicationOncreate =%s,
ActivityOnCreate    =%s,
AppBeginLoading     =%s,
AppShowUI           =%s,
AppShowImageView    =%s,
AppWindowVisible    =%s,
AppQueueIdleEvent    =%s,

AppServiceConnected =%s,
LBSAgentBegin       =%s,
LBSAgentEnd         =%s,
InitOkHttpBegin     =%s,
InitOkHttpEnd       =%s,
HttpRequestTook     =%s,
]
`,
		coldStartInfo.pid,
		coldStartInfo.LinuxBroadcast,
		coldStartInfo.ApplicationOnCreate,
		coldStartInfo.ActivityOnCreate,
		coldStartInfo.AppBeginLoading,
		coldStartInfo.AppShowUI,
		coldStartInfo.AppShowImageView,
		coldStartInfo.AppWindowVisible,
		coldStartInfo.AppQueueIdleEvent,
		coldStartInfo.AppServiceConnected,
		coldStartInfo.LBSAgentBegin,
		coldStartInfo.LBSAgentEnd,
		coldStartInfo.InitOkHttpBegin,
		coldStartInfo.InitOkHttpEnd,
		coldStartInfo.HttpRequestTookMs,
	)
}

type AppColdStartTAG uint16

const (
	LinuxBroadcast AppColdStartTAG = iota
	ApplicationOnCreate
	ApplicationOnCreateEnd
	ActivityOnCreate
	ActivityOnCreateEnd
	AppBeginLoading
	AppShowUI
	AppShowImageView
	AppWindowVisible
	AppQueueIdleEvent
	HttpRequestTook
	UnKnow
)

func analyseColdStartFile(filepath *string, sequence int, fileNum int) *list.List {
	data, err := ioutil.ReadFile(*filepath)
	if err != nil {
		fmt.Println("File parse failed, ", err)
		return nil
	}

	content := string(data)
	// fmt.Println("read file:\n", content)
	sSlice := strings.Split(content, "\n")

	coldStartInfoList := list.New()

	for _, line := range sSlice {
		if len(line) < 18 {
			continue
		}
		handleLine(line, coldStartInfoList)

	}
	return coldStartInfoList
}

var emtpy time.Time

func handleLine(line string, coldStartInfoList *list.List) (time.Time, AppColdStartTAG) {
	// fmt.Println("handle line = ", line)
	timestamp := log_time + line[0:18]
	timeRes, _ := time.Parse("2006-01-02 15:04:05", timestamp)

	var tag AppColdStartTAG = UnKnow
	if strings.Contains(line, linuxBroadcast) {
		tag = LinuxBroadcast
		appColdStartInfo = &AppColdStartInfo{LowMemKillList: list.New(), BackgroundGCList: list.New()}
		coldStartInfoList.PushBack(appColdStartInfo)
		appColdStartInfo.LinuxBroadcast = timeRes
	} else if strings.Contains(line, applicationOnCreate) {
		tag = ApplicationOnCreate
		if appColdStartInfo != nil {
			appColdStartInfo.ApplicationOnCreate = timeRes
		}
	} else if strings.Contains(line, applicationOnCreateEnd) {
		tag = ApplicationOnCreateEnd
		if appColdStartInfo != nil {
			appColdStartInfo.ApplicationOnCreateEnd = timeRes
		}
	} else if strings.Contains(line, activityOnCreate) {
		tag = ActivityOnCreate
		if appColdStartInfo != nil {
			appColdStartInfo.ActivityOnCreate = timeRes
		}
	} else if strings.Contains(line, activityOnCreateEnd) {
		tag = ActivityOnCreateEnd
		if appColdStartInfo != nil {
			appColdStartInfo.ActivityOnCreateEnd = timeRes
		}
	} else if strings.Contains(line, appBeginLoading) {
		tag = AppBeginLoading
		if appColdStartInfo != nil {
			appColdStartInfo.AppBeginLoading = timeRes
		}
	} else if strings.Contains(line, appShowUI) {
		tag = AppShowUI
		if appColdStartInfo != nil {
			appColdStartInfo.AppShowUI = timeRes
			// appColdStartInfo = nil
		}
	} else if strings.Contains(line, appShowImageView) {
		tag = AppShowImageView
		if appColdStartInfo != nil && appColdStartInfo.AppShowImageView.Day() <= emtpy.Day() {
			appColdStartInfo.AppShowImageView = timeRes
			// appColdStartInfo = nil
		}
	} else if strings.Contains(line, appWindowVisible) {
		tag = AppWindowVisible
		if appColdStartInfo != nil {
			appColdStartInfo.AppWindowVisible = timeRes
		}
	} else if strings.Contains(line, queueIdleEvent) {
		tag = AppQueueIdleEvent
		if appColdStartInfo != nil {
			appColdStartInfo.AppQueueIdleEvent = timeRes
			// appColdStartInfo = nil
		}
	} else if strings.Contains(line, lowMemoryKiller) {
		if appColdStartInfo != nil {
			appColdStartInfo.LowMemKillList.PushBack(line)
		}
	} else if strings.Contains(line, initOkhttp) {
		if appColdStartInfo != nil {
			if appColdStartInfo.InitOkHttpBegin != nil {
				appColdStartInfo.InitOkHttpEnd = &timeRes
			} else {
				appColdStartInfo.InitOkHttpBegin = &timeRes
			}
		}
	} else if strings.Contains(line, appServiceConnected) {
		if appColdStartInfo != nil {
			appColdStartInfo.AppServiceConnected = timeRes
		}
	} else if strings.Contains(line, backgroundGC) {
		if appColdStartInfo != nil && strings.Contains(line, strconv.Itoa(appColdStartInfo.pid)) {
			appColdStartInfo.BackgroundGCList.PushBack(line)
		}
	} else if strings.Contains(line, "Start proc") {
		if appColdStartInfo != nil {
			result := regl.FindAllStringSubmatch(line, -1)
			if len(result) >= 1 {
				result = pidRegl.FindAllStringSubmatch(result[0][0], -1)
				appColdStartInfo.pid, _ = strconv.Atoi(result[0][0])
			}
		}
	} else if strings.Contains(line, httpRequestTook) {
		if appColdStartInfo != nil {
			appColdStartInfo.HttpRequestTookMs = line
		}
	} else if strings.Contains(line, lbsAgentbegin) {
		if appColdStartInfo != nil {
			appColdStartInfo.LBSAgentBegin = &timeRes
		}
	} else if strings.Contains(line, lbsAgentEnd) {
		if appColdStartInfo != nil {
			appColdStartInfo.LBSAgentEnd = &timeRes
		}
	}

	return timeRes, tag
}
