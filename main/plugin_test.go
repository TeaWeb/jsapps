package main

import (
	"github.com/TeaWeb/jsapps/probes"
	"testing"
)

func TestPlugin(t *testing.T) {
	engine := probes.NewScriptEngine()
	err := engine.RunScript(`(function () {
	var probe = new ProcessProbe();
	probe.name = "MongoDB";
	probe.commandName = "mongod"
	probe.commandPatterns = [ "/mongod" ];
	probe.commandVersion = "${commandFile} --version"
	probe.onProcess(function (process) {
		console.log("process name:", JSON.stringify(process));
		return false;
	});
	probe.onParseVersion(function (v) {
		var match = v.match(/v[0-9\.]+/);
		if (!match) {
			return "not matched";
		}
		return match[0];
	});
	probe.run();
})();

`)
	if err != nil {
		t.Fatal("ERROR:", err)
	}

	t.Log("======")
	for _, app := range engine.Apps() {
		t.Log(app.Name, "version:"+app.Version, "cmd:"+app.Cmdline)
	}
}

func TestLoadJs(t *testing.T) {
	loadJs("/Users/liuxiangchao/Documents/Projects/pp/apps/TeaWebGit/build/src/main/configs/jsapps.js")
}

func TestElasticSearch(t *testing.T) {
	engine := probes.NewScriptEngine()
	t.Log(engine.RunScript(`var probe = new ProcessProbe();
            probe.name = "ElasticSearch";
            probe.site = "https://www.elastic.co";
            probe.docSite = "https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html";
            probe.developer = "Elasticsearch B.V.";
            probe.commandName = "java";
            probe.commandPatterns = [ "Elasticsearch" ];
			probe.commandVersion = "${commandFile} --version"
            probe.onProcess(function (p) {
				var args = parseArgs(p.cmdline)
				var homeDir = ""; 
				for (var i = 0; i < args.length; i ++) {
					var arg = args[i];
					var index = arg.indexOf("-Des.path.home=");
					if (index < 0) {
						continue;
					}
					homeDir = arg.substring("-Des.path.home=".length);
				}
				if (homeDir.length > 0) {
					p.dir = homeDir;
					p.file = homeDir + "/bin/elasticsearch";
				}
                return true;
            });
			probe.onParseVersion(function (v) {
				return v;
			});
            probe.run();`))

	apps := engine.Apps()
	if len(apps) > 0 {
		t.Log("apps:", apps)
		t.Log("version:", apps[0].Version)
	}
}
