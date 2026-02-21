package main

import (
	"fmt"
)

type ColumnLayout struct {
	NumWidth  int
	AddWidth  int
	DelWidth  int
	FileWidth int

	TitleWidth      int
	AuthorWidth     int
	HeadRefWidth    int

	ShowFiles  bool
	ShowDate   bool
	ShowTitle  bool
	ShowAuthor bool
}

type droppableFixed struct {
	name  string
	width int
}

type variableCol struct {
	name string
	min  int
}

func calculateLayout(prs []PullRequest, opt options) ColumnLayout {
	if len(prs) == 0 {
		return ColumnLayout{NumWidth: 4, ShowFiles: true, ShowDate: true, ShowTitle: true, ShowAuthor: true}
	}

	maxNum := 4
	maxAdd := 1
	maxDel := 1
	maxFile := 1
	for _, pr := range prs {
		if w := len(fmt.Sprintf("%d", pr.Number)); w > maxNum {
			maxNum = w
		}
		if w := len(fmt.Sprintf("%d", pr.Additions)); w > maxAdd {
			maxAdd = w
		}
		if w := len(fmt.Sprintf("%d", pr.Deletions)); w > maxDel {
			maxDel = w
		}
		if w := len(fmt.Sprintf("%d", pr.ChangedFiles)); w > maxFile {
			maxFile = w
		}
	}

	// Natural widths for variable columns
	natW := map[string]int{
		"authorName":  0,
		"title":       0,
		"headRefName": 0,
	}
	for _, pr := range prs {
		if w := displayWidth(pr.AuthorName); w > natW["authorName"] {
			natW["authorName"] = w
		}
		if w := displayWidth(pr.Title); w > natW["title"] {
			natW["title"] = w
		}
		if w := displayWidth(pr.HeadRefName); w > natW["headRefName"] {
			natW["headRefName"] = w
		}
	}

	// Required fixed width: "#N  " + "+N/-N"
	required := (maxNum + 3) + (maxAdd + maxDel + 3)

	droppableFixedCols := []droppableFixed{
		{name: "files", width: maxFile + 10}, // "  N files  "
		{name: "date", width: 22},            // "  " + 20 chars
	}

	variableCols := []variableCol{
		{name: "title", min: 15},
		{name: "authorName", min: 6},
		{name: "headRefName", min: 12},
	}

	show := map[string]bool{
		"files":      true,
		"date":       true,
		"title":      true,
		"authorName": true,
	}

	colW := map[string]int{}

	effWidth := termWidth() - fzfMargin(opt)
	baseAvail := effWidth - required

	computeAvail := func() int {
		a := baseAvail
		for _, df := range droppableFixedCols {
			if show[df.name] {
				a -= df.width
			}
		}
		visCount := 0
		for _, v := range variableCols {
			if s, ok := show[v.name]; ok && s || !ok {
				visCount++
			}
		}
		a -= visCount * 2 // separators
		return a
	}

	tryFit := func(avail int, cols []variableCol) bool {
		natTotal := 0
		for _, c := range cols {
			natTotal += natW[c.name]
		}
		if natTotal <= avail {
			for _, c := range cols {
				colW[c.name] = natW[c.name]
			}
			return true
		}
		for i, v := range cols {
			others := 0
			for j, c := range cols {
				if i == j {
					continue
				}
				if w, ok := colW[c.name]; ok {
					others += w
				} else {
					others += natW[c.name]
				}
			}
			thisW := avail - others
			if thisW < v.min {
				thisW = v.min
			}
			if thisW > natW[v.name] {
				thisW = natW[v.name]
			}
			colW[v.name] = thisW

			total := 0
			for _, c := range cols {
				if w, ok := colW[c.name]; ok {
					total += w
				} else {
					total += natW[c.name]
				}
			}
			if total <= avail {
				for j := i + 1; j < len(cols); j++ {
					if _, ok := colW[cols[j].name]; !ok {
						colW[cols[j].name] = natW[cols[j].name]
					}
				}
				return true
			}
		}
		return false
	}

	// Phase 1 & 2
	if !tryFit(computeAvail(), variableCols) {
		// Phase 3: set all variable columns to min width
		for _, v := range variableCols {
			colW[v.name] = v.min
		}

		type droppable struct {
			name     string
			isFixed  bool
		}
		droppableAll := []droppable{
			{name: "files", isFixed: true},
			{name: "date", isFixed: true},
			{name: "title", isFixed: false},
			{name: "authorName", isFixed: false},
		}

		for _, drop := range droppableAll {
			// Check if current state fits
			var visibleVar []variableCol
			for _, v := range variableCols {
				if s, ok := show[v.name]; ok && s || !ok {
					visibleVar = append(visibleVar, v)
				}
			}
			avail := computeAvail()
			total := 0
			for _, v := range visibleVar {
				total += colW[v.name]
			}
			if total <= avail {
				break
			}

			show[drop.name] = false
			if !drop.isFixed {
				delete(colW, drop.name)
			}

			// Re-run Phase 2 on remaining visible variable columns
			visibleVar = nil
			for _, v := range variableCols {
				if s, ok := show[v.name]; ok && s || !ok {
					visibleVar = append(visibleVar, v)
				}
			}
			if tryFit(computeAvail(), visibleVar) {
				break
			}
		}
	}

	layout := ColumnLayout{
		NumWidth:  maxNum,
		AddWidth:  maxAdd,
		DelWidth:  maxDel,
		FileWidth: maxFile,
		ShowFiles: show["files"],
		ShowDate:  show["date"],
		ShowTitle: show["title"],
		ShowAuthor: show["authorName"],
	}
	if w, ok := colW["title"]; ok {
		layout.TitleWidth = w
	}
	if w, ok := colW["authorName"]; ok {
		layout.AuthorWidth = w
	}
	if w, ok := colW["headRefName"]; ok {
		layout.HeadRefWidth = w
	}

	return layout
}
