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
	apps := loadJs("/Users/liuxiangchao/Documents/Projects/pp/apps/TeaWebGit/build/src/main/configs/jsapps.js")
	for _, app := range apps {
		t.Log(app.Name)
	}
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


func TestPhpFpm(t *testing.T) {
	engine := probes.NewScriptEngine()
	err := engine.RunScript(`(function () {
	var probe = new ProcessProbe(); // 构造对象
	probe.author = ""; // 探针作者
	probe.id = "local_1545123416213326756"; // 探针ID，
	probe.name = "Redis"; // App名称
	probe.site = "https://redis.io/"; // App官方网站
	probe.docSite = "https://redis.io/documentation"; // 官方文档网址
	probe.developer = "redislabs"; // App开发者公司、团队或者个人名称
	probe.commandName = "redis-server"; // App启动的命令名称
	probe.commandPatterns = [""]; // 进程匹配规则
	probe.commandVersion = "{commandFile} -v"; // 获取版本信息的命令

	// 进程筛选
	probe.onProcess(function (p) {
		return true;
	});

	// 版本信息分析
	probe.onParseVersion(function (v) {
		return v;
	});

	// 运行探针
	probe.run();
})()`)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("apps:", engine.Apps()[0])
}