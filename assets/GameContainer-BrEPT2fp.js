import{a as n,w as g,y as w,v as k,p as s,O as b}from"./chunk-OIYGIGL5-CDuitOYr.js";import{O as C}from"./Options-BvmdAYDC.js";/**
 * @license lucide-react v0.548.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const y=t=>t.replace(/([a-z0-9])([A-Z])/g,"$1-$2").toLowerCase(),j=t=>t.replace(/^([A-Z])|[\s-_]+(\w)/g,(e,o,a)=>a?a.toUpperCase():o.toLowerCase()),f=t=>{const e=j(t);return e.charAt(0).toUpperCase()+e.slice(1)},v=(...t)=>t.filter((e,o,a)=>!!e&&e.trim()!==""&&a.indexOf(e)===o).join(" ").trim(),N=t=>{for(const e in t)if(e.startsWith("aria-")||e==="role"||e==="title")return!0};/**
 * @license lucide-react v0.548.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */var L={xmlns:"http://www.w3.org/2000/svg",width:24,height:24,viewBox:"0 0 24 24",fill:"none",stroke:"currentColor",strokeWidth:2,strokeLinecap:"round",strokeLinejoin:"round"};/**
 * @license lucide-react v0.548.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const A=n.forwardRef(({color:t="currentColor",size:e=24,strokeWidth:o=2,absoluteStrokeWidth:a,className:i="",children:r,iconNode:m,...d},h)=>n.createElement("svg",{ref:h,...L,width:e,height:e,stroke:t,strokeWidth:a?Number(o)*24/Number(e):o,className:v("lucide",i),...!r&&!N(d)&&{"aria-hidden":"true"},...d},[...m.map(([u,c])=>n.createElement(u,c)),...Array.isArray(r)?r:[r]]));/**
 * @license lucide-react v0.548.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const x=(t,e)=>{const o=n.forwardRef(({className:a,...i},r)=>n.createElement(A,{ref:r,iconNode:e,className:v(`lucide-${y(f(t))}`,`lucide-${t}`,a),...i}));return o.displayName=f(t),o};/**
 * @license lucide-react v0.548.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const O=[["path",{d:"M15 21v-8a1 1 0 0 0-1-1h-4a1 1 0 0 0-1 1v8",key:"5wwlr5"}],["path",{d:"M3 10a2 2 0 0 1 .709-1.528l7-6a2 2 0 0 1 2.582 0l7 6A2 2 0 0 1 21 10v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z",key:"r6nss1"}]],E=x("house",O);/**
 * @license lucide-react v0.548.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */const R=[["path",{d:"M9.671 4.136a2.34 2.34 0 0 1 4.659 0 2.34 2.34 0 0 0 3.319 1.915 2.34 2.34 0 0 1 2.33 4.033 2.34 2.34 0 0 0 0 3.831 2.34 2.34 0 0 1-2.33 4.033 2.34 2.34 0 0 0-3.319 1.915 2.34 2.34 0 0 1-4.659 0 2.34 2.34 0 0 0-3.32-1.915 2.34 2.34 0 0 1-2.33-4.033 2.34 2.34 0 0 0 0-3.831A2.34 2.34 0 0 1 6.35 6.051a2.34 2.34 0 0 0 3.319-1.915",key:"1i5ecw"}],["circle",{cx:"12",cy:"12",r:"3",key:"1v7zrd"}]],$=x("settings",R),S=g(function(){const{gameID:e}=w(),o=k(),[a,i]=n.useState(!1),r=n.useRef(null),m=()=>{o("/")},d=()=>{o(`/game/${e}`)},h=()=>{i(!0)},u=()=>{i(!1)};return n.useEffect(()=>{console.log("Requesting wake lock");const c=async()=>{if("wakeLock"in navigator)try{const l=await navigator.wakeLock.request("screen");r.current=l}catch(l){console.error("Failed to acquire wake lock:",l)}};c();const p=()=>{document.visibilityState==="visible"&&r.current===null&&c()};return document.addEventListener("visibilitychange",p),()=>{document.removeEventListener("visibilitychange",p),r.current&&r.current.release().catch(l=>{console.error("Failed to release wake lock:",l)})}},[]),s.jsxs("div",{className:"h-screen overflow-hidden",children:[s.jsx("div",{className:"h-[calc(100vh-52px)] overflow-hidden pt-6",children:s.jsx(b,{})}),s.jsxs("div",{className:"fixed bottom-0 left-0 w-full bg-[#0D111A]/90 backdrop-blur-md border-t border-gray-700 flex items-center justify-between px-3 py-3 z-50",children:[s.jsx("div",{className:"flex items-center gap-5",children:s.jsx("button",{onClick:m,className:"text-gray-300 hover:text-white transition-transform hover:scale-110 active:scale-95","aria-label":"Home",children:s.jsx(E,{size:26})})}),s.jsx("div",{className:"text-white font-semibold tracking-widest text-lg select-none cursor-pointer",onClick:d,children:e}),s.jsx("button",{onClick:h,className:"text-gray-300 hover:text-white transition-transform hover:scale-110 active:scale-95","aria-label":"Settings",children:s.jsx($,{size:26})})]}),a&&s.jsx("div",{className:"fixed inset-0 bg-black/70 backdrop-blur-sm flex items-center justify-center z-50",onClick:u,children:s.jsx("div",{className:"bg-[#0D111A] border border-gray-700 rounded-lg p-8 max-w-md w-full mx-4",onClick:c=>c.stopPropagation(),children:s.jsx(C,{onClose:u})})})]})});export{S as default};
