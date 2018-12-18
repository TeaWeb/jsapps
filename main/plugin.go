package main

import (
	"fmt"
	"github.com/TeaWeb/jsapps/probes"
	"github.com/TeaWeb/plugin/apps"
	"github.com/TeaWeb/plugin/loader"
	"github.com/TeaWeb/plugin/plugins"
	"log"
	"os"
)

func main() {
	p := plugins.NewPlugin()
	p.Name = "Javascript Apps"
	p.Description = "使用Javascript语言实现的本地服务探测"
	p.Version = "v0.0.1"
	p.Developer = "TeaWeb"
	p.Site = "https://github.com/TeaWeb/build"
	p.Code = "jsapps.teaweb"
	p.Date = "2018-12-16"

	// 从.js中读取
	p.OnReloadApps(func() {
		results := loadJs("")
		if len(results) > 0 {
			p.AddApp(results ...)
		}

		log.Println("[jsapps]found " + fmt.Sprintf("%d", len(results)) + " apps")
	})
	p.ReloadApps()

	loader.Start(p)
}

func loadJs(jsFile string) (results []*apps.App) {
	if len(jsFile) == 0 {
		cwd, _ := os.Getwd()
		jsFile = cwd + string(os.PathSeparator) + "configs" + string(os.PathSeparator) + "jsapps.js"
	}

	parser := probes.NewParser(jsFile)
	_, o, err := parser.LoadFunctions()
	if err != nil {
		return nil
	}
	for _, key := range o.Keys() {
		v, err := o.Get(key)
		if err != nil {
			log.Println("[jsapps]" + err.Error())
			return
		}
		engine := probes.NewScriptEngine()
		err = engine.RunScript("(" + v.String() + ")()")
		if err != nil {
			log.Println("[jsapps]" + err.Error())
			continue
		} else {
			results = append(results, engine.Apps() ...)
		}
	}
	return
}
