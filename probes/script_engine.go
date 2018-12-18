package probes

import (
	"errors"
	"github.com/TeaWeb/plugin/apps"
	"github.com/TeaWeb/plugin/apps/probes"
	"github.com/TeaWeb/plugin/utils/types"
	"github.com/iwind/TeaGo/maps"
	"github.com/robertkrimen/otto"
	"log"
	"reflect"
	"sync"
)

type ScriptEngine struct {
	vm      *otto.Otto
	locker  sync.Mutex
	results []*apps.App
}

func NewScriptEngine() *ScriptEngine {
	engine := &ScriptEngine{
		vm: otto.New(),
	}
	engine.init()
	return engine
}

func (this *ScriptEngine) init() {
	this.vm.Set("runProcessProbe", this.runProcessProbe)
	this.vm.Set("parseArgs", this.parseArgs)

	_, err := this.vm.Run(`function ProcessProbe () {
	this.id = "";
	this.author = "";
    this.name = "";
    this.site = "";
    this.docSite = "";
    this.developer = "";
    this.commandName = "";
    this.commandPatterns = [];
    this.commandVersion = "";

	this.processFilter = null;
	this.versionParser = null;

	this.onProcess = function (processFilter) {
		this.processFilter = processFilter;
	};

    this.onParseVersion = function (versionParser) {
		if (typeof(versionParser) != "function") {
			throw new Error('onParseVersion() must accept a valid function');
		}
		this.versionParser = versionParser;
    };

    this.run = function () {
		return runProcessProbe(this, {
			"name": this.name,
			"site": this.site,
			"docSite": this.docSite,
			"developer": this.developer,
			"commandName": this.commandName,
			"commandPatterns": this.commandPatterns,
			"commandVersion": this.commandVersion,
			"processFilter": this.processFilter,
			"versionParser": this.versionParser
		});
    };
}`)
	if err != nil {
		log.Println("[jsapps]" + err.Error())
	}
}

func (this *ScriptEngine) RunScript(script string) error {
	_, err := this.vm.Run(script)
	if err != nil {
		return err
	}
	return nil
}

func (this *ScriptEngine) Apps() []*apps.App {
	return this.results
}

func (this *ScriptEngine) throw(err error) {
	if err != nil {
		value, _ := this.vm.Call("new Error", nil, err.Error())
		panic(value)
	}
}

func (this *ScriptEngine) runProcessProbe(call otto.FunctionCall) otto.Value {
	v := call.Argument(1)
	m, err := v.Export()
	if err != nil {
		this.throw(err)
	} else {
		attr := maps.NewMap(m)

		probe := probes.NewProcessProbe()
		probe.Author = attr.GetString("author")
		probe.Name = attr.GetString("name")
		probe.Site = attr.GetString("site")
		probe.DocSite = attr.GetString("docSite")
		probe.Developer = attr.GetString("developer")
		probe.CommandName = attr.GetString("commandName")
		probe.CommandVersion = attr.GetString("commandVersion")

		if len(probe.CommandName) == 0 {
			this.throw(errors.New("'commandName' should not be empty"))
		}

		// patterns
		patterns := attr.Get("commandPatterns")
		t := reflect.TypeOf(patterns)
		if t != nil {
			if t.Kind() == reflect.Slice {
				value := reflect.ValueOf(patterns)
				count := value.Len()
				for i := 0; i < count; i ++ {
					pattern := value.Index(i).Interface()
					if pattern != nil {
						patternString := types.String(pattern)
						if len(patternString) > 0 {
							probe.CommandPatterns = append(probe.CommandPatterns, patternString)
						}
					}
				}
			}
		}

		// process filter
		{
			f, err := v.Object().Get("processFilter")
			if err != nil {
				this.throw(err)
			}
			if f.IsFunction() {
				probe.OnProcess(func(process *apps.Process) bool {
					arg0 := maps.Map{
						"name":      process.Name,
						"pid":       process.Pid,
						"ppid":      process.Ppid,
						"cwd":       process.Cwd,
						"user":      process.User,
						"uid":       process.Uid,
						"gid":       process.Gid,
						"cmdline":   process.Cmdline,
						"file":      process.File,
						"dir":       process.Dir,
						"isRunning": process.IsRunning,
					}
					result, err := f.Call(call.Argument(0), arg0)

					// 获取新的值
					process.File = arg0.GetString("file")
					process.Dir = arg0.GetString("dir")

					if err != nil {
						this.throw(err)
					}
					if result.IsBoolean() {
						b, _ := result.ToBoolean()
						return b
					}
					return true
				})
			}
		}

		// version parser
		{
			f, err := v.Object().Get("versionParser")
			if err != nil {
				this.throw(err)
			}
			if f.IsFunction() {
				probe.OnParseVersion(func(versionString string) (string, error) {
					result, err := f.Call(call.Argument(0), versionString)
					if err != nil {
						this.throw(err)
					}
					if result.IsString() {
						return result.ToString()
					}
					return "", nil
				})
			}
		}

		resultApps, err := probe.Run()
		if err != nil {
			this.throw(err)
		} else {
			this.locker.Lock()
			this.results = append(this.results, resultApps ...)
			this.locker.Unlock()
		}
	}

	return otto.Value{}
}

func (this *ScriptEngine) parseArgs(call otto.FunctionCall) otto.Value {
	arg0 := call.Argument(0)
	if !arg0.IsString() {
		return otto.Value{}
	}
	s, _ := arg0.ToString()
	args := apps.ParseArgs(s)
	v, _ := this.vm.ToValue(args)
	return v
}
