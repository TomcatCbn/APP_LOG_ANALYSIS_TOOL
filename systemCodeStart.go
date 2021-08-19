package main

import (
	"container/list"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
)

// var appColdStartInfo *AppColdStartInfo

// var regl = regexp.MustCompile(startProc)
// var pidRegl = regexp.MustCompile(`\d{4,5}`)

var lowmemorykillerShowNum = 0

func getSystemColdStart(filePath *string) *list.List {
	return analyseSystemColdStartFile(filePath, 0, 0)
}

func analyseSystemColdStartFile(filepath *string, sequence int, fileNum int) *list.List {
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
		handleLineSys(line, coldStartInfoList)

	}

	fmt.Println("---------------- lowmemeorykiller show " + strconv.Itoa(lowmemorykillerShowNum) + " times ---------------------")
	return coldStartInfoList
}

func handleLineSys(line string, coldStartInfoList *list.List) {
	// fmt.Println("handle line = ", line)
	// timestamp := log_time + line[0:18]
	// _, _ := time.Parse("2006-01-02 15:04:05", timestamp)

	if strings.Contains(line, lowMemoryKiller) {
		fmt.Println(line)
		lowmemorykillerShowNum += 1
	} else if strings.Contains(line, backgroundGC) {
		fmt.Println(line)
	} else if strings.Contains(line, "Start proc") {
		fmt.Println(line)
	}
}
