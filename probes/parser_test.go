package probes

import (
	"github.com/TeaWeb/plugin/apps/probes"
	"testing"
)

func TestParser_AddProbe(t *testing.T) {
	parser := NewParser("/Users/liuxiangchao/Documents/Projects/pp/apps/TeaWebGit/build/src/main/configs/jsapps.js")
	probe := new(probes.ProcessProbe)
	probe.Name = "Test"
	err := parser.AddProbe(probe)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParser_Parse(t *testing.T) {
	parser := NewParser("/Users/liuxiangchao/Documents/Projects/pp/apps/TeaWebGit/build/src/main/configs/jsapps.js")
	t.Log(parser.Parse())
}

func TestParser_RemoveProbe(t *testing.T) {
	parser := NewParser("/Users/liuxiangchao/Documents/Projects/pp/apps/TeaWebGit/build/src/main/configs/jsapps.js")
	t.Log(parser.RemoveProbe("local_1545048811647795669"))
}
