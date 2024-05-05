package util

import "github.com/microcosm-cc/bluemonday"

func XSSRemover(str string, doc ...bool) string {
	p := bluemonday.NewPolicy()
	//添加允许的标签
	if len(doc) > 0 {
		if doc[0] {
			p.AllowElements("p", "span", "li", "strong", "ul", "em", "u", "ol", "br")
			p.AllowAttrs("style").OnElements("p", "span", "li")
		}
	}

	return p.Sanitize(str)
}
