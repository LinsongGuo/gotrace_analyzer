package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-gota/gota/dataframe"
)

type ParsedEvent struct {
	Type      string
	Goroutine uint64
	Processor int
	Timestamp int64
	// LinkType  string
	// LinkTs    int64
	Stack []StackFrame
}

type StackFrame struct {
	Func string
	File string
	Line int
}

func analyze(path, bench string, gonum int) []string {

	tracefile := fmt.Sprintf("%s_%d.trace", bench, gonum)
	tracepath := filepath.Join(path, tracefile)
	trace, err := os.Open(tracepath)

	// fmt.Println("dfsfd")
	// fmt.Println(path, bench, gonum)

	res, err := Parse(trace, "")
	if err != nil {
		panic(err)
	}

	var stuff []ParsedEvent
	for _, event := range res.Events {
		eventType := EventDescriptions[event.Type]

		// linkType := "NoLink"
		// var linkTs int64 = 0
		// if event.Link != nil {
		// 	linkType = EventDescriptions[event.Link.Type].Name
		// 	linkTs = event.Link.Ts
		// }

		thing := ParsedEvent{
			Type:      eventType.Name,
			Timestamp: event.Ts,
			Processor: event.P,
			Goroutine: event.G,
			// LinkType:  linkType,
			// LinkTs:    linkTs,
		}
		stk := res.Stacks[event.StkID]
		for _, frame := range stk {
			thing.Stack = append(thing.Stack, StackFrame{
				File: frame.File,
				Func: frame.Fn,
				Line: frame.Line,
			})
		}
		// fmt.Printf("P: %d\n", event.P)
		stuff = append(stuff, thing)
	}

	jsonfile := fmt.Sprintf("%s_%d.json", bench, gonum)
	jsonpath := filepath.Join(path, jsonfile)
	data, err := json.MarshalIndent(stuff, "", "  ")
	os.WriteFile(jsonpath, data, 0660)

	cnt := 0
	overhead := []string{} // []string{strconv.Itoa(gonum)}
	var preempt_stuff []ParsedEvent
	for idx, event := range stuff {
		// fmt.Printf("idx: %d\n", idx)
		if event.Type == "GoPreempt" {
			cnt += 1
			preempt_stuff = append(preempt_stuff, event)

			for next := idx + 1; next < len(stuff); next++ {
				if stuff[next].Processor == event.Processor {
					stuff[next].Timestamp -= event.Timestamp
					o := strconv.FormatFloat(float64(stuff[next].Timestamp)/1000.0, 'f', 3, 32)
					overhead = append(overhead, o)
					preempt_stuff = append(preempt_stuff, stuff[next])
					break
				}
			}
			// stuff[idx+1].Timestamp -= event.Timestamp
			// o := strconv.FormatFloat(float64(stuff[idx+1].Timestamp)/1000.0, 'f', 3, 32)
			// overhead = append(overhead, o)
			// preempt_stuff = append(preempt_stuff, stuff[idx+1])
		}
	}

	jsonfile2 := fmt.Sprintf("%s_preempt_%d.json", bench, gonum)
	jsonpath2 := filepath.Join(path, jsonfile2)
	data2, err := json.MarshalIndent(preempt_stuff, "", "  ")
	os.WriteFile(jsonpath2, data2, 0660)

	fmt.Printf("preempt %d\n", cnt)
	return overhead
}

func writecsv(data [][]string, path, bench string) {

	filename := fmt.Sprintf("%s.csv", bench)
	filepath := filepath.Join(path, filename)

	maxnum := len(data)
	maxlen := 0
	for i := 0; i < maxnum; i++ {
		if maxlen < len(data[i]) {
			maxlen = len(data[i])
		}
	}

	header := []string{}
	for i := 1; i <= len(data); i++ {
		header = append(header, strconv.Itoa(i))
	}

	data2 := [][]string{header}
	for i := 0; i < maxlen; i++ {
		row := []string{}
		for j := 0; j < maxnum; j++ {
			if i < len(data[j]) {
				row = append(row, data[j][i])
			} else {
				row = append(row, "")
			}
		}
		data2 = append(data2, row)
	}

	// Create a new DataFrame
	df := dataframe.LoadRecords(data2)

	// Write the DataFrame to a CSV file
	f, _ := os.Create(filepath)
	err := df.WriteCSV(f)
	if err != nil {
		fmt.Println("Error writing CSV:", err)
	}
}

func main() {
	args := os.Args[1:]
	// if len(args) == 2 {
	// 	fmt.Println("Usage: go run main.go <path> <age>")
	// 	return
	// }

	path := args[0]
	bench := args[1]

	analyze(path, bench, 1)

	// result := [][]string{}
	// for gonum := 1; gonum <= 10; gonum++ {
	// 	tmp := analyze(path, bench, gonum)
	// 	result = append(result, tmp)
	// 	fmt.Println(tmp)
	// }

	// writecsv(result, path, bench)
}
