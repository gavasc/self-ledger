

export const index = 0;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/fallbacks/layout.svelte.js')).default;
export const universal = {
  "ssr": false,
  "prerender": true
};
export const universal_id = "src/routes/+layout.ts";
export const imports = ["_app/immutable/nodes/0.BjpH32vv.js","_app/immutable/chunks/Ct-7OXHN.js","_app/immutable/chunks/CYeplY8F.js","_app/immutable/chunks/BMifteI8.js"];
export const stylesheets = [];
export const fonts = [];
