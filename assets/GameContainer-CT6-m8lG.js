import{a as i,w as f,y as g,v,p as s,O as w}from"./chunk-OIYGIGL5-CDuitOYr.js";import{O as b}from"./Options-BvmdAYDC.js";/**
 * @license lucide-react v0.548.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const C=t=>t.replace(/([a-z0-9])([A-Z])/g,"$1-$2").toLowerCase(),j=t=>t.replace(/^([A-Z])|[\s-_]+(\w)/g,(e,o,r)=>r?r.toUpperCase():o.toLowerCase()),p=t=>{const e=j(t);return e.charAt(0).toUpperCase()+e.slice(1)},u=(...t)=>t.filter((e,o,r)=>!!e&&e.trim()!==""&&r.indexOf(e)===o).join(" ").trim(),k=t=>{for(const e in t)if(e.startsWith("aria-")||e==="role"||e==="title")return!0};/**
 * @license lucide-react v0.548.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */var y={xmlns:"http://www.w3.org/2000/svg",width:24,height:24,viewBox:"0 0 24 24",fill:"none",stroke:"currentColor",strokeWidth:2,strokeLinecap:"round",strokeLinejoin:"round"};/**
 * @license lucide-react v0.548.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const N=i.forwardRef(({color:t="currentColor",size:e=24,strokeWidth:o=2,absoluteStrokeWidth:r,className:n="",children:a,iconNode:d,...c},l)=>i.createElement("svg",{ref:l,...y,width:e,height:e,stroke:t,strokeWidth:r?Number(o)*24/Number(e):o,className:u("lucide",n),...!a&&!k(c)&&{"aria-hidden":"true"},...c},[...d.map(([m,x])=>i.createElement(m,x)),...Array.isArray(a)?a:[a]]));/**
 * @license lucide-react v0.548.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const h=(t,e)=>{const o=i.forwardRef(({className:r,...n},a)=>i.createElement(N,{ref:a,iconNode:e,className:u(`lucide-${C(p(t))}`,`lucide-${t}`,r),...n}));return o.displayName=p(t),o};/**
 * @license lucide-react v0.548.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const A=[["path",{d:"M15 21v-8a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v8",key:"5wwlr5"}],["path",{d:"M3 10a2 2 0 0 1 .709-1.528l7-6a2 2 0 0 1 2.582 0l7 6A2 2 0 0 1 21 10v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z",key:"r6nss1"}]],O=h("house",A);/**
 * @license lucide-react v0.548.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const $=[["path",{d:"M9.671 4.136a2.34 2.34 0 0 1 4.659 0 2.34 2.34 0 0 0 3.319 1.915 2.34 2.34 0 0 1 2.33 4.033 2.34 2.34 0 0 0 0 3.831 2.34 2.34 0 0 1-2.33 4.033 2.34 2.34 0 0 0-3.319 1.915 2.34 2.34 0 0 1-4.659 0 2.34 2.34 0 0 0-3.32-1.915 2.34 2.34 0 0 1-2.33-4.033 2.34 2.34 0 0 0 0-3.831A2.34 2.34 0 0 1 6.35 6.051a2.34 2.34 0 0 0 3.319-1.915",key:"1i5ecw"}],["circle",{cx:"12",cy:"12",r:"3",key:"1v7zrd"}]],z=h("settings",$),P=f(function(){const{gameID:e}=g(),o=v(),[r,n]=i.useState(!1),a=()=>{o("/")},d=()=>{o(`/game/${e}`)},c=()=>{n(!0)},l=()=>{n(!1)};return s.jsxs("div",{className:"h-screen overflow-hidden",children:[s.jsx("div",{className:"h-[calc(100vh-52px)] overflow-hidden pt-6",children:s.jsx(w,{})}),s.jsxs("div",{className:"fixed bottom-0 left-0 w-full bg-[#0D111A]/90 backdrop-blur-md border-t border-gray-700 flex items-center justify-between px-3 py-3 z-50",children:[s.jsx("div",{className:"flex items-center gap-5",children:s.jsx("button",{onClick:a,className:"text-gray-300 hover:text-white transition-transform hover:scale-110 active:scale-95","aria-label":"Home",children:s.jsx(O,{size:26})})}),s.jsx("div",{className:"text-white font-semibold tracking-widest text-lg select-none cursor-pointer",onClick:d,children:e}),s.jsx("button",{onClick:c,className:"text-gray-300 hover:text-white transition-transform hover:scale-110 active:scale-95","aria-label":"Settings",children:s.jsx(z,{size:26})})]}),r&&s.jsx("div",{className:"fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50",onClick:l,children:s.jsx("div",{className:"bg-[#0D111A] border border-gray-700 rounded-lg p-8 max-w-md w-full mx-4",onClick:m=>m.stopPropagation(),children:s.jsx(b,{onClose:l})})})]})});export{P as default};
