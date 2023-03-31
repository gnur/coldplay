package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"path"
	"reflect"
	"sort"
	"strings"
	"time"
)

var templateFunctions = template.FuncMap{
	"percent": func(a, b int) float64 {
		return float64(a) / float64(b) * 100
	},
	"safeHTML": func(s interface{}) template.HTML {
		return template.HTML(fmt.Sprint(s))
	},
	"crop": func(s string, i int) string {
		if len(s) > i {
			return s[0:(i-2)] + "..."
		}
		return s
	},
	"hasSuffix": func(s, suffix string) bool {
		return strings.HasSuffix(s, suffix)
	},
	"filename": func(f string) string {
		return path.Base(f)
	},
	"filesize": func(b int64) string {
		if b == 0 {
			return "unknown"
		}
		const unit = 1024
		if b < unit {
			return fmt.Sprintf("%d B", b)
		}
		div, exp := int64(unit), 0
		for n := b / unit; n >= unit; n /= unit {
			div *= unit
			exp++
		}
		return fmt.Sprintf("%.1f %ciB",
			float64(b)/float64(div), "KMGTPE"[exp])
	},
	"floatDecimal": func(f float64) string {
		return fmt.Sprintf("%.2f", f)
	},
	"nl2br": func(str string) template.HTML {
		return template.HTML(strings.Replace(str, "\n", "<br />", -1))
	},
	"add": func(b, a interface{}) (interface{}, error) {
		av := reflect.ValueOf(a)
		bv := reflect.ValueOf(b)

		switch av.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			switch bv.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return av.Int() + bv.Int(), nil
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return av.Int() + int64(bv.Uint()), nil
			case reflect.Float32, reflect.Float64:
				return float64(av.Int()) + bv.Float(), nil
			default:
				return nil, fmt.Errorf("add: unknown type for %q (%T)", bv, b)
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			switch bv.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return int64(av.Uint()) + bv.Int(), nil
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return av.Uint() + bv.Uint(), nil
			case reflect.Float32, reflect.Float64:
				return float64(av.Uint()) + bv.Float(), nil
			default:
				return nil, fmt.Errorf("add: unknown type for %q (%T)", bv, b)
			}
		case reflect.Float32, reflect.Float64:
			switch bv.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				return av.Float() + float64(bv.Int()), nil
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				return av.Float() + float64(bv.Uint()), nil
			case reflect.Float32, reflect.Float64:
				return av.Float() + bv.Float(), nil
			default:
				return nil, fmt.Errorf("add: unknown type for %q (%T)", bv, b)
			}
		default:
			return nil, fmt.Errorf("add: unknown type for %q (%T)", av, a)
		}
	},
	"Iterate": func(offset, limit int64, results int64) [][2]int64 {
		var i int64
		var Items [][2]int64
		for i = 0; i <= (results / limit); i++ {
			Items = append(Items, [2]int64{
				i + 1,
				i * limit,
			})
		}

		itemIndices := []int{0, 1, 2}

		for i, b := range Items {
			if b[1] == offset {
				itemIndices = append(itemIndices, i-1, i, i+1)
				break
			}
		}
		if len(Items) > 8 {
			l := len(Items)
			itemIndices = append(itemIndices, l-3, l-2, l-2)
		}

		properIndices := func(a []int) []int {

			var indices []int
			seen := make(map[int]bool)

			for _, b := range a {
				if _, ok := seen[b]; !ok {
					indices = append(indices, b)
					seen[b] = true
				}
			}

			sort.Ints(indices)

			return indices
		}(itemIndices)

		var retItems [][2]int64

		var lastPage int64
		lastPage = 0
		for _, i := range properIndices {
			if i < 0 || i >= len(Items) {
				continue
			}
			if lastPage+1 != Items[i][0] {
				retItems = append(retItems, [2]int64{-1, -1})
			}

			retItems = append(retItems, Items[i])
			lastPage = Items[i][0]
		}

		return retItems

	},
	"prettyTime": func(s interface{}) template.HTML {
		t, ok := s.(time.Time)
		if !ok {
			return ""
		}
		if t.IsZero() {
			return template.HTML("never")
		}
		return template.HTML(t.Format("2006-01-02 15:04:05"))
	},
	"page": func(dir, q string, offset, limit int64) template.URL {
		v := url.Values{}
		v.Add("q", q)
		v.Add("l", fmt.Sprintf("%v", limit))
		if dir == "next" {
			start := offset + limit
			v.Add("o", fmt.Sprintf("%v", start))
		} else {
			start := offset - limit
			if start > 0 {
				v.Add("o", fmt.Sprintf("%v", start))
			}
		}
		return template.URL(v.Encode())

	},
	"json": func(s interface{}) template.HTML {
		json, _ := json.MarshalIndent(s, "", "  ")
		return template.HTML(strings.Replace(string(json), "\n", "<br />", -1))
	},
	"relativeTime": func(s interface{}) template.HTML {
		t, ok := s.(time.Time)
		if !ok {
			return ""
		}
		if t.IsZero() {
			return template.HTML("never")
		}
		tense := "ago"
		diff := time.Since(t)
		seconds := int64(diff.Seconds())
		if seconds < 0 {
			tense = "from now"
		}
		var quantifier string

		if seconds < 60 {
			quantifier = "s"
		} else if seconds < 3600 {
			quantifier = "m"
			seconds /= 60
		} else if seconds < 86400 {
			quantifier = "h"
			seconds /= 3600
		} else if seconds < 604800 {
			quantifier = "d"
			seconds /= 86400
		} else if seconds < 31556736 {
			quantifier = "w"
			seconds /= 604800
		} else {
			quantifier = "y"
			seconds /= 31556736
		}

		return template.HTML(fmt.Sprintf("%v%s %s", seconds, quantifier, tense))
	},
}
