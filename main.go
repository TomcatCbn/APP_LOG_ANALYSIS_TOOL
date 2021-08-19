package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const MEM_TEMPLATE_TOTAL_PSS = "Total PSS by process"
const MEM_TEMPLATE_TOTAL_RAM = "Total RAM"
const MEM_TEMPLATE_TOTAL_FREE_RAM = "Free RAM"
const MEM_TEMPLATE_TOTAL_USED_RAM = "Used RAM"

type TEMPLATE_MODE uint8

const (
	MODE_UNKNOW TEMPLATE_MODE = iota
	MODE_PROCESS
	MODE_TOTAL_RAM
)

var totalRAMLine []string

func main() {
	fmt.Print(
		`
**************************************************************************
	This is a tool, which helps to analyse meminfo and cold start.       
	Please pass params mode to decide which analysis mode you want to use         
**************************************************************************
`,
	)
	fMode := flag.Int("m", 0, "0 = meminfo, 1 = app cold start")
	fptr := flag.String("f", "", "file path to read from")
	dptr := flag.String("d", "", "dir path to read from")
	packptr := flag.String("k", "", "filter package name in meminfo")
	flag.Parse()
	fmt.Println("value of file path is", *fptr)
	fmt.Println("value of dir path is", *dptr)
	fmt.Println("value of filter key word is", *packptr)
	switch *fMode {
	case 0:
		analyseMemEntry(fptr, dptr, packptr)
	case 1:
		analyseColdStart(fptr, dptr)
	case 2:
		analyseSystem(fptr)
	default:
	}
}

func analyseSystem(fptr *string) {
	getSystemColdStart(fptr)
}

func analyseColdStart(fptr *string, dptr *string) {

	if *dptr != "" {
		dir, err := os.Stat(*dptr)
		if err != nil {
			fmt.Println("not dir, ", *dptr)
			return
		}

		if !dir.IsDir() {
			fmt.Println("not dir, ", *dptr)
			return
		}

		var fileInfo []os.FileInfo
		fileInfo, err = ioutil.ReadDir(*dptr)
		if err != nil {
			fmt.Println("dir error, ", err)
		}

		for index, v := range fileInfo {
			tar := *dptr + v.Name()
			if strings.HasSuffix(tar, ".log") || strings.HasSuffix(tar, ".txt") {
				analyseColdStartEntry(&tar, index)
			}

		}
	} else if *fptr != "" {
		analyseColdStartEntry(fptr, 0)
	}

}

func analyseColdStartEntry(file *string, fileIndex int) {
	infoList := getAppColdStart(file)
	fmt.Println("cold start analyse...")
	var index int = 1
	for infos := infoList.Front(); infos != nil; infos = infos.Next() {
		var info *AppColdStartInfo = infos.Value.(*AppColdStartInfo)
		fmt.Println()
		fmt.Println("-----------", *file, "----------------")
		fmt.Println("-------------", index, "---------------")
		index++
		fmt.Println(info)

		fmt.Printf("linux_broadcast->application_oncreate = %d\n", (info.ApplicationOnCreate.UnixNano()-info.LinuxBroadcast.UnixNano())/1e6)
		fmt.Printf("application_oncreate->activity_oncreate = %d\n", (info.ActivityOnCreate.UnixNano()-info.ApplicationOnCreate.UnixNano())/1e6)
		fmt.Printf("activity_oncreate->app_begin_loading = %d\n", (info.AppBeginLoading.UnixNano()-info.ApplicationOnCreate.UnixNano())/1e6)
		fmt.Printf("app_begin_loading->app_show_ui = %d\n", (info.AppShowUI.UnixNano()-info.AppBeginLoading.UnixNano())/1e6)
		fmt.Printf("linux_broadcast->app_show_ui = %d\n", (info.AppShowUI.UnixNano()-info.LinuxBroadcast.UnixNano())/1e6)
		fmt.Printf("linux_broadcast->app_show_image_view = %d\n", (info.AppShowImageView.UnixNano()-info.LinuxBroadcast.UnixNano())/1e6)
		fmt.Printf("linux_broadcast->app_window_visibile = %d\n", (info.AppWindowVisible.UnixNano()-info.LinuxBroadcast.UnixNano())/1e6)
		fmt.Printf("linux_broadcast->app_queue_idle_event = %d\n", (info.AppQueueIdleEvent.UnixNano()-info.LinuxBroadcast.UnixNano())/1e6)

		fmt.Println("other key point")
		fmt.Println("app service connected = ", (info.AppServiceConnected.UnixNano()-info.LinuxBroadcast.UnixNano())/1e6)
		if info.LBSAgentBegin != nil && info.LBSAgentEnd != nil {
			fmt.Println("app lbsagent = ", (info.LBSAgentEnd.UnixNano()-info.LBSAgentBegin.UnixNano())/1e6)
		}
		if info.InitOkHttpBegin != nil && info.InitOkHttpEnd != nil {
			fmt.Println("app init okhttp = ", (info.InitOkHttpEnd.UnixNano()-info.InitOkHttpBegin.UnixNano())/1e6)
		}

		fmt.Println()

		if info.LowMemKillList.Len() > 0 {
			fmt.Println("low memory killer appears")
			for lowKiller := info.LowMemKillList.Front(); lowKiller != nil; lowKiller = lowKiller.Next() {
				fmt.Println(lowKiller.Value)
			}
		}

		fmt.Println()

		if info.BackgroundGCList.Len() > 0 {
			fmt.Println("background gc appears")
			for gc := info.BackgroundGCList.Front(); gc != nil; gc = gc.Next() {
				fmt.Println(gc.Value)
			}
		}

	}
}

func analyseMemEntry(fptr *string, dptr *string, packptr *string) {
	// map
	var memRecordMap = make(map[string]([]*ProcessInfo))

	if *dptr != "" {
		dir, err := os.Stat(*dptr)
		if err != nil {
			fmt.Println("not dir, ", *dptr)
			return
		}

		if !dir.IsDir() {
			fmt.Println("not dir, ", *dptr)
			return
		}

		var fileInfo []os.FileInfo
		fileInfo, err = ioutil.ReadDir(*dptr)
		if err != nil {
			fmt.Println("dir error, ", err)
		}
		fileNum := len(fileInfo)

		for index, v := range fileInfo {
			tar := *dptr + v.Name()
			analyseMemFile(&tar, memRecordMap, index, fileNum)
		}
	} else if *fptr != "" {
		analyseMemFile(fptr, memRecordMap, 0, 1)
	}

	// analyse mem done

	for key, value := range memRecordMap {
		fmt.Println("packageName=", key)
		for _, proc := range value {
			if proc != nil {
				fmt.Println("memSize = ", proc.Memeory)
			} else {
				fmt.Println("memSize = empty")
			}
		}
		fmt.Println()
	}

	for index, val := range totalRAMLine {
		fmt.Println(val)
		if index%4 == 3 {
			fmt.Println()
			fmt.Println()
		}

	}

	if *packptr != "" {
		fmt.Println("print fillter = ", *packptr)
		for key, value := range memRecordMap {
			if strings.Contains(key, *packptr) {
				fmt.Println("packageName=", key)
				for _, proc := range value {
					if proc != nil {
						fmt.Println("memSize = ", proc.Memeory)
					} else {
						fmt.Println("memSize = empty")
					}
				}
				fmt.Println()
			}
		}
	}
}

func analyseMemFile(filepath *string, recordMap map[string]([]*ProcessInfo), sequence int, fileNum int) {
	data, err := ioutil.ReadFile(*filepath)
	if err != nil {
		fmt.Println("File parse failed, ", err)
		return
	}
	fmt.Println("now analyse file, ", *filepath)

	content := string(data)
	// fmt.Println("read file:\n", content)
	sSlice := strings.Split(content, "\n")

	analyseMemInfo(sSlice, recordMap, sequence, fileNum)
}

func analyseMemInfo(lines []string, recordMap map[string]([]*ProcessInfo), sequence int, fileNume int) {
	if len(lines) == 0 {
		return
	}

	var mode = MODE_UNKNOW
	for _, line := range lines {
		// fmt.Println("cur line = ", line)
		if len(line) <= 1 {
			continue
		}
		if strings.HasPrefix(line, " ") {
			switch mode {
			case MODE_PROCESS:
				processInfo := analyseProcess(line)
				if s := recordMap[processInfo.PackageName]; s != nil {
					s[sequence] = processInfo
				} else {
					s = make([]*ProcessInfo, fileNume)
					recordMap[processInfo.PackageName] = s
					s[sequence] = processInfo
				}
			case MODE_TOTAL_RAM:
				analyseTotal(line)
			case MODE_UNKNOW:
			default:
			}
		} else {
			if strings.HasPrefix(line, MEM_TEMPLATE_TOTAL_PSS) {
				mode = MODE_PROCESS
			} else if strings.HasPrefix(line, MEM_TEMPLATE_TOTAL_RAM) {
				mode = MODE_TOTAL_RAM
			} else {
				mode = MODE_UNKNOW
			}
		}

	}

}

func analyseProcess(line string) *ProcessInfo {
	parts := strings.Split(strings.Trim(line, " "), ": ")
	//xx,xxK
	var memStr = parts[0]
	processInfo := &ProcessInfo{}
	if memory, err := handleKilobytesStr(memStr); err != nil {
		fmt.Println("handleKilobytesStr fail, so 0. ", err)
	} else {
		processInfo.Memeory = memory
	}

	parts = strings.Split(strings.Trim(parts[1], " "), " ")
	processInfo.PackageName = parts[0]

	// fmt.Println("processInfo:", processInfo)

	return processInfo
}

func analyseTotal(line string) {
	// fmt.Println("analyseTotal:", line)
	totalRAMLine = append(totalRAMLine, line)
}

func handleKilobytesStr(kilobytes string) (uint32, error) {
	if kilobytes == "" {
		return 0, nil
	}
	parts := strings.Split(kilobytes, ",")
	size := len(parts)
	if size <= 1 {
		res, err := strconv.Atoi(parts[0][:len(parts[0])-1])
		return uint32(res), err
	} else {
		left := parts[0]
		right := parts[1][:len(parts[1])-1]
		res, err := strconv.Atoi(left + right)
		return uint32(res), err
	}
}

type ProcessInfo struct {
	PackageName string
	Memeory     uint32
}

func (proc *ProcessInfo) String() string {
	return fmt.Sprintf("[%s, %d]", proc.PackageName, proc.Memeory)
}

type SystemInfo struct {
	FreeRAM      uint32
	CachedKernel uint32
	FreeForApp   uint32
}

func (sys *SystemInfo) String() string {
	return fmt.Sprintf("[%d, %d, %d]", sys.FreeRAM, sys.CachedKernel, sys.FreeForApp)
}
