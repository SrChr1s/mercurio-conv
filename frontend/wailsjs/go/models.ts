export namespace convert {
	
	export class Options {
	    format: string;
	    quality: number;
	    maxWidth: number;
	    maxHeight: number;
	
	    static createFrom(source: any = {}) {
	        return new Options(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.format = source["format"];
	        this.quality = source["quality"];
	        this.maxWidth = source["maxWidth"];
	        this.maxHeight = source["maxHeight"];
	    }
	}

}

export namespace main {
	
	export class FileInfo {
	    id: string;
	    name: string;
	    size: number;
	    format: string;
	    width: number;
	    height: number;
	    thumb: string;
	
	    static createFrom(source: any = {}) {
	        return new FileInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.size = source["size"];
	        this.format = source["format"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.thumb = source["thumb"];
	    }
	}
	export class Request {
	    ids: string[];
	    options: convert.Options;
	    formats: Record<string, string>;
	    output: string;
	    outDir: string;
	    suffix: string;
	
	    static createFrom(source: any = {}) {
	        return new Request(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ids = source["ids"];
	        this.options = this.convertValues(source["options"], convert.Options);
	        this.formats = source["formats"];
	        this.output = source["output"];
	        this.outDir = source["outDir"];
	        this.suffix = source["suffix"];
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

