package common

import (
	"encoding/json"
	"encoding/xml"
	"strings"
)

func Xml2Json(src []byte) []byte {
	type xmlNode struct {
		XMLName xml.Name
		Attrs   []xml.Attr `xml:",any,attr"`
		Content string     `xml:",chardata"`
		Nodes   []xmlNode  `xml:",any"`
	}
	rootDTO := xmlNode{}
	if err := xml.Unmarshal(src, &rootDTO); err != nil {
		return []byte(err.Error())
	}

	var walk func(node xmlNode) any
	walk = func(node xmlNode) any {
		items := map[string]any{"!": "[" + node.XMLName.Space + "]: " + node.XMLName.Local}
		if c := strings.TrimSpace(node.Content); c != "" {
			items["%"] = c
		}
		for _, a := range node.Attrs {
			items["$"+a.Name.Local] = a.Value
		}
		if len(node.Nodes) != 0 {
			children := []any{}
			for _, c := range node.Nodes {
				children = append(children, walk(c))
			}
			items["&"] = children
		}
		return items
	}

	j, err := json.Marshal(walk(rootDTO))
	if err != nil {
		return []byte(err.Error())
	}
	return j
}
