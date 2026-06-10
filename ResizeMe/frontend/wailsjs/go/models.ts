export namespace main {
	
	export class Preset {
	    id: string;
	    name: string;
	    width: number;
	    height: number;
	
	    static createFrom(source: any = {}) {
	        return new Preset(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.width = source["width"];
	        this.height = source["height"];
	    }
	}
	export class Config {
	    presets: Preset[];
	    activePresetId: string;
	    centerAfterResize: boolean;
	    hotkey: string;
	    autoStart: boolean;
	    firstRun: boolean;
	    loadError?: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.presets = this.convertValues(source["presets"], Preset);
	        this.activePresetId = source["activePresetId"];
	        this.centerAfterResize = source["centerAfterResize"];
	        this.hotkey = source["hotkey"];
	        this.autoStart = source["autoStart"];
	        this.firstRun = source["firstRun"];
	        this.loadError = source["loadError"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

