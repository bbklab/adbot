package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"

	"github.com/bbklab/adbot/pkg/label"
)

func main() {
	bs, _ := ioutil.ReadFile("window_dump.xml")
	nodes, err := parseAndroidUINodes(bs)
	if err != nil {
		logrus.Fatalln(err)
	}

	for _, node := range nodes {
		fmt.Println(node.ResourceID, node.Text, node.ContentDesc, node.Bounds)
	}
}

// AndroidUINode is exported
type AndroidUINode struct {
	Index       string `json:"index"`        // 4
	Text        string `json:"text"`         // 手机电池已充满(100%)
	ResourceID  string `json:"resource_id"`  // com.android.systemui:id/clear_all_button
	Package     string `json:"package"`      // com.android.systemui
	ContentDesc string `json:"content_desc"` // 清除所有通知。
	Bounds      string `json:"bounds"`       // [301,1138][419,1256]
	XY          [2]int `json:"xy"`           // MiddleXY() of UINode
}

// MiddleXY is exported
func (n *AndroidUINode) MiddleXY() (int, int, error) {
	bounds := strings.NewReplacer(
		[]string{
			",", " ",
			"[", " ",
			"]", " ",
		}...).Replace(string(n.Bounds))

	locations := strings.Fields(bounds)
	if len(locations) != 4 {
		return -1, -1, errors.New("the UI bounds are invalid")
	}

	x1, y1, x2, y2 := locations[0], locations[1], locations[2], locations[3]
	if x1 == "" || y1 == "" || x2 == "" || y2 == "" {
		return -1, -1, errors.New("can't find the UI bounds value")
	}

	x1N, _ := strconv.Atoi(x1)
	y1N, _ := strconv.Atoi(y1)
	x2N, _ := strconv.Atoi(x2)
	y2N, _ := strconv.Atoi(y2)
	x := (x1N + x2N) / 2
	y := (y1N + y2N) / 2

	return x, y, nil
}

func parseAndroidUINodes(xmldata []byte) ([]*AndroidUINode, error) {
	var (
		nodes        = make([]string, 0, 0)
		noderx       = regexp.MustCompile(`<node ([^>]*)/>`)
		nodesMatched = noderx.FindAllSubmatch(xmldata, -1)
	)

	for _, nodeMatched := range nodesMatched {
		if len(nodeMatched) >= 2 {
			nodes = append(nodes, string(nodeMatched[1]))
		}
	}

	var (
		nodekvrx = regexp.MustCompile(`([^ \t]+)="([^"]+)"`)
		ret      = []*AndroidUINode{}
	)
	for _, node := range nodes {
		nodekvsMatched := nodekvrx.FindAllSubmatch([]byte(node), -1)
		lbs := label.Labels{}
		for _, nodekvMatched := range nodekvsMatched {
			if len(nodekvMatched) >= 3 {
				key := string(nodekvMatched[1])
				val := string(nodekvMatched[2])
				lbs.Set(key, val)
			}
		}
		uinode := &AndroidUINode{ // here we only pick up the fields we want
			Index:       lbs.Get("index"),
			Text:        lbs.Get("text"),
			ResourceID:  lbs.Get("resource-id"),
			Package:     lbs.Get("package"),
			ContentDesc: lbs.Get("content-desc"),
			Bounds:      lbs.Get("bounds"),
		}
		x, y, _ := uinode.MiddleXY()
		uinode.XY = [2]int{x, y}
		ret = append(ret, uinode)
	}

	return ret, nil
}
