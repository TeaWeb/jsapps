package probes

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/TeaWeb/plugin/apps/probes"
	"github.com/robertkrimen/otto"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"
)

// 分析器定义
type Parser struct {
	jsFile string
}

// 取得分析器对象
func NewParser(jsFile string) *Parser {
	return &Parser{
		jsFile: jsFile,
	}
}

// 获取所有的指针信息
func (this *Parser) Parse() ([]map[string]interface{}, error) {
	_, o, err := this.LoadFunctions()
	if err != nil {
		return nil, err
	}
	results := []map[string]interface{}{}
	for _, key := range o.Keys() {
		v, err := o.Get(key)
		if err != nil {
			return nil, err
		}
		s := regexp.MustCompile("(\\w+)\\.run\\(\\)").ReplaceAllString(v.String(), "return $1")
		vm := otto.New()
		vm.Run(`function ProcessProbe() {
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
}`)
		result, err := vm.Run("(" + s + ")()")
		if err != nil {
			return nil, err
		}
		probe := map[string]interface{}{}
		o := result.Object()
		for _, key := range []string{"id", "name", "developer", "site", "docSite", "commandName", "commandPatterns", "commandVersion"} {
			v, _ := o.Get(key)
			if v.IsObject() {
				exportedValue, err := v.Export()
				if err == nil {
					probe[key] = exportedValue
				} else {
					return nil, err
				}
			} else if v.IsString() {
				probe[key] = v.String()
			}
		}
		results = append(results, probe)
	}
	return results, nil
}

// 添加Probe
func (this *Parser) AddProbe(probe *probes.ProcessProbe) error {
	content, o, err := this.LoadFunctions()
	if err != nil {
		return err
	}

	funcs := []string{}
	for _, key := range o.Keys() {
		f, _ := o.Get(key)
		funcs = append(funcs, f.String())
	}

	funcId := fmt.Sprintf("local_%d", time.Now().UnixNano())
	template := `function () {
	var probe = new ProcessProbe(); // 构造对象
	probe.author = ""; // 探针作者
	probe.id = ${ID}; // 探针ID，
	probe.name = ${NAME}; // App名称
	probe.site = ${SITE}; // App官方网站
	probe.docSite = ${DOC_SITE}; // 官方文档网址
	probe.developer = ${DEVELOPER}; // App开发者公司、团队或者个人名称
	probe.commandName = ${COMMAND_NAME}; // App启动的命令名称
	probe.commandPatterns = ${COMMAND_PATTERNS}; // 进程匹配规则
	probe.commandVersion = ${COMMAND_VERSION}; // 获取版本信息的命令

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
}`
	if len(probe.CommandPatterns) == 0 {
		probe.CommandPatterns = []string{}
	}
	template = strings.NewReplacer(
		"${ID}", this.toJSON(funcId),
		"${NAME}", this.toJSON(probe.Name),
		"${SITE}", this.toJSON(probe.Site),
		"${DOC_SITE}", this.toJSON(probe.DocSite),
		"${DEVELOPER}", this.toJSON(probe.Developer),
		"${COMMAND_NAME}", this.toJSON(probe.CommandName),
		"${COMMAND_PATTERNS}", this.toJSON(probe.CommandPatterns),
		"${COMMAND_VERSION}", this.toJSON(probe.CommandVersion)).
		Replace(template)

	_, _, err = otto.Run("var f = " + template)
	if err != nil {
		return errors.New(err.Error() + ":" + template)
	}

	funcs = append(funcs, template)

	content = regexp.MustCompile("\"probes\":\\s*\\[(.|\n)*]").ReplaceAllLiteralString(content, "\"probes\": ["+strings.Join(funcs, ",\n")+"]")

	return ioutil.WriteFile(this.jsFile, []byte(content), 0666)
}

// 添加Probe函数
func (this *Parser) AddProbeFunc(probeFunc string) error {
	content, o, err := this.LoadFunctions()
	if err != nil {
		return err
	}

	funcs := []string{}
	for _, key := range o.Keys() {
		f, _ := o.Get(key)
		funcs = append(funcs, f.String())
	}

	funcs = append(funcs, probeFunc)

	content = regexp.MustCompile("\"probes\":\\s*\\[(.|\n)*]").ReplaceAllLiteralString(content, "\"probes\": ["+strings.Join(funcs, ",\n")+"]")

	return ioutil.WriteFile(this.jsFile, []byte(content), 0666)
}

// 删除Probe
func (this *Parser) RemoveProbe(probeId string) error {
	if len(probeId) == 0 {
		return errors.New("'probeId' should not be empty")
	}
	content, o, err := this.LoadFunctions()
	if err != nil {
		return err
	}

	funcs := []string{}
	for _, key := range o.Keys() {
		f, _ := o.Get(key)
		s := f.String()
		if strings.Index(s, "\""+probeId+"\"") > 0 {
			continue
		}
		funcs = append(funcs, s)
	}

	content = regexp.MustCompile("\"probes\":\\s*\\[(.|\n)*]").ReplaceAllLiteralString(content, "\"probes\": ["+strings.Join(funcs, ",\n")+"]")

	return ioutil.WriteFile(this.jsFile, []byte(content), 0666)
}

// 替换Probe
func (this *Parser) ReplaceProbe(probeId string, funcString string) error {
	if len(probeId) == 0 {
		return errors.New("'probeId' should not be empty")
	}

	content, o, err := this.LoadFunctions()
	if err != nil {
		return err
	}

	funcs := []string{}
	for _, key := range o.Keys() {
		f, _ := o.Get(key)
		s := f.String()
		if strings.Index(s, "\""+probeId+"\"") > 0 {
			funcs = append(funcs, funcString)
			continue
		}
		funcs = append(funcs, s)
	}

	content = regexp.MustCompile("\"probes\":\\s*\\[(.|\n)*]").ReplaceAllLiteralString(content, "\"probes\": ["+strings.Join(funcs, ",\n")+"]")

	return ioutil.WriteFile(this.jsFile, []byte(content), 0666)
}

// 查找Probe
func (this *Parser) FindProbeFunction(probeId string) (string, error) {
	if len(probeId) == 0 {
		return "", errors.New("'probeId' should not be empty")
	}
	_, o, err := this.LoadFunctions()
	if err != nil {
		return "", err
	}

	for _, key := range o.Keys() {
		f, _ := o.Get(key)
		s := f.String()
		if strings.Index(s, "\""+probeId+"\"") > 0 {
			return s, nil
		}
	}
	return "", errors.New("function not found")
}

// 加载函数数据
func (this *Parser) LoadFunctions() (content string, funcs *otto.Object, err error) {
	data, err := ioutil.ReadFile(this.jsFile)
	if err != nil {
		if os.IsNotExist(err) {
			// 创建
			content = `var ENGINE = {
    "version": 0,
    "probes": []
};`
			err = ioutil.WriteFile(this.jsFile, []byte(content), 0666)
			if err != nil {
				return "", nil, err
			}
			data = []byte(content)
		} else {
			return "", nil, err
		}
	}
	content = string(data)
	vm := otto.New()
	_, err = vm.Run(content)
	if err != nil {
		return content, nil, err
	}
	engine, err := vm.Get("ENGINE")
	if err != nil {
		return content, nil, err
	}
	if !engine.IsObject() {
		return content, nil, errors.New("invalid 'ENGINE' value")
	}
	v, err := engine.Object().Get("probes")
	if err != nil {
		return content, nil, err
	}
	return content, v.Object(), nil
}

func (this *Parser) toJSON(o interface{}) string {
	data, err := json.Marshal(o)
	if err != nil {
		return ""
	}
	return string(data)
}
