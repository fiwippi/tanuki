(()=>{function n(t){t.directive("intersect",(i,{expression:r,modifiers:c},{evaluateLater:o,cleanup:s})=>{let d=o(r),e=new IntersectionObserver(l=>{l.forEach(a=>{a.intersectionRatio!==0&&(d(),c.includes("once")&&e.disconnect())})});e.observe(i),s(()=>{e.disconnect()})})}document.addEventListener("alpine:init",()=>{window.Alpine.plugin(n)});})();