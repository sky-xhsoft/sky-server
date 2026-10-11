package volcengine

import (
	"encoding/json"
	"strings"
	"testing"
)

// 火山控制台「新建任务」实际提交的参数（实时同传、原文 ASR 实时流式），经本系统解析后再序列化，关键字段不能丢
func TestSpeechTaskRule_RoundTripKeepsConsoleFields(t *testing.T) {
	cases := map[string]string{
		"实时同传": `{"Source":{"Url":"rtmp://a/live/src"},"Output":{"Url":["rtmp://a/live/dst"]},
			"Extra":{"AMix":{"AudioEnhance":true,"BGVolume":0.2},"Separator":{"HandleType":1}},"Delay":0,
			"RealTime":{"VoiceShow":true,"Separator":true},
			"Subtitle":{"Position":{"MarginLr":0.2,"MarginTb":0.05},
				"Original":{"FontColor":"#ffffff","Border":{"Color":"#000000@0.7","W":1},"Box":{"Color":"#000000@0.7","W":1,"Filter":0.7},"MaxCharNumber":60,"FontSize":40,"Font":"arial","Language":"zh","MaxRowNumber":0,"Hidden":false,"ShowPriority":0},
				"GlossaryMap":{"火山引擎":"Volcengine"},"ReplaceMap":{"Volcano":"Volcengine"},
				"Translation":[{"FontColor":"#ffffff","Border":{"Color":"#4EA6FF@0.5","W":1},"Box":{"Color":"#000000@0.7","W":1,"Filter":0.7},"MaxCharNumber":60,"FontSize":40,"Font":"arial","Language":"en","MaxRowNumber":0,"Hidden":false,"ShowPriority":100}]},
			"Pipe":"realtime","StartTime":"2026-10-11T15:16:07+08:00","EndTime":"2026-10-14T15:16:07+08:00","Factory":{"TTS":"","SID":""}}`,
		"原文ASR": `{"Source":{"Url":"rtmp://a/live/src"},"Output":{"Url":["rtmp://a/live/dst"]},"Extra":{},"Delay":0,
			"Subtitle":{"Position":{"MarginLr":0.2,"MarginTb":0.05},"Original":{"Language":"zh","MaxRowNumber":2,"Hidden":false,"ShowPriority":0},
				"ReplaceMap":{"Volcano":"Volcengine"},"HotWords":["火山引擎","声影同传"],"Translation":[]},
			"Pipe":"realtime","StartTime":"x","EndTime":"y","Factory":{"TTS":"","SID":""}}`,
	}
	for name, in := range cases {
		var rule SpeechTaskRule
		if err := json.Unmarshal([]byte(in), &rule); err != nil {
			t.Fatalf("%s 解析失败：%v", name, err)
		}
		out, _ := json.Marshal(rule)
		s := string(out)
		want := []string{`"Delay":0`, `"MaxRowNumber"`, `"ShowPriority":0`, `"Hidden":false`, `"Factory":{"TTS":"","SID":""}`, `"Translation":[`}
		if name == "实时同传" {
			want = append(want, `"Separator":true`, `"BGVolume":0.2`, `"HandleType":1`, `"Filter":0.7`, `"GlossaryMap":{"火山引擎":"Volcengine"}`, `"#4EA6FF@0.5"`)
		} else {
			want = append(want, `"HotWords":["火山引擎","声影同传"]`, `"Translation":[]`)
		}
		for _, w := range want {
			if !strings.Contains(s, w) {
				t.Errorf("%s 序列化后缺少 %s：\n%s", name, w, s)
			}
		}
	}
}
