export namespace main {
	
	export class Settings {
	    selectedTheme: string;
	    defaultUsername: string;
	    defaultHost: string;
	    defaultPort: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.selectedTheme = source["selectedTheme"];
	        this.defaultUsername = source["defaultUsername"];
	        this.defaultHost = source["defaultHost"];
	        this.defaultPort = source["defaultPort"];
	    }
	}

}

