export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set(["favicon.png","svelte.svg","tauri.svg","vite.svg"]),
	mimeTypes: {".png":"image/png",".svg":"image/svg+xml"},
	_: {
		client: {start:"_app/immutable/entry/start.f01w80Ey.js",app:"_app/immutable/entry/app.D8aliflM.js",imports:["_app/immutable/entry/start.f01w80Ey.js","_app/immutable/chunks/DQYwDHLu.js","_app/immutable/chunks/CYeplY8F.js","_app/immutable/chunks/B-lcsz3Q.js","_app/immutable/entry/app.D8aliflM.js","_app/immutable/chunks/CYeplY8F.js","_app/immutable/chunks/BjjmIcVZ.js","_app/immutable/chunks/Ct-7OXHN.js","_app/immutable/chunks/B-lcsz3Q.js","_app/immutable/chunks/BCzA7BCX.js","_app/immutable/chunks/BMifteI8.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./nodes/0.js')),
			__memo(() => import('./nodes/1.js')),
			__memo(() => import('./nodes/2.js'))
		],
		remotes: {
			
		},
		routes: [
			{
				id: "/",
				pattern: /^\/$/,
				params: [],
				page: { layouts: [0,], errors: [1,], leaf: 2 },
				endpoint: null
			}
		],
		prerendered_routes: new Set([]),
		matchers: async () => {
			
			return {  };
		},
		server_assets: {}
	}
}
})();
