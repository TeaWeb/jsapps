function ProcessProbe() {
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
            "site": this.size,
            "docSite": this.docSite,
            "developer": this.developer,
            "commandName": this.commandName,
            "commandPatterns": this.commandPatterns,
            "commandVersion": this.commandVersion,
            "processFilter": this.processFilter,
            "versionParser": this.versionParser
        });
    };
}