// Apply the saved theme and text size (boot.js) before styles first apply.
// A theme requested by the page's address takes precedence. Waiting for
// app.js at the foot of the page let the header paint in the system theme
// and default size first, then change as the page loaded.
(function () {
  const white = [255, 255, 255], black = [0, 0, 0];
  const mix = (a, b, p) => a.map((v, i) => v + (b[i] - v) * p);
  const hex = (c) => "#" + c.map((v) => Math.round(v).toString(16).padStart(2, "0")).join("");
  const lum = (c) => c.reduce((sum, v, i) => {
    v /= 255;
    return sum + [0.2126, 0.7152, 0.0722][i] * (v <= 0.03928 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4);
  }, 0);
  const contrast = (a, b) => { const x = lum(a), y = lum(b); return (Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05); };
  const vars = (c, dark) => {
    const bg = dark ? [26, 26, 30] : white, target = dark ? [35, 35, 39] : [244, 244, 246];
    let a = c;
    for (let p = 0; p <= 1 && contrast(a, target) < 4.5; p += 0.02) a = mix(c, dark ? white : black, p);
    const fg = contrast(a, [20, 20, 26]) > contrast(a, white) ? [20, 20, 26] : white;
    return `--accent:${hex(a)};--accent-fg:${hex(fg)};--accent-soft:${hex(mix(bg, a, dark ? .14 : .09))};--sel:${hex(mix(bg, a, dark ? .2 : .11))}`;
  };
  window.paintAccent = function (value) {
    const color = /^#[\da-f]{6}$/i.test(value || "") ? value : "#574ab8";
    const c = [1, 3, 5].map((i) => parseInt(color.slice(i, i + 2), 16));
    let style = document.getElementById("accentCSS");
    if (!style) { style = document.createElement("style"); style.id = "accentCSS"; document.head.appendChild(style); }
    const light = vars(c, false), dark = vars(c, true);
    style.textContent = `html:not(.omarchy):root{${light}}@media(prefers-color-scheme:dark){html:not(.omarchy):root:not([data-theme="light"]){${dark}}}html:not(.omarchy):root[data-theme="dark"]{${dark}}`;
  };
  window.paintAccent(window.bootPrefs && window.bootPrefs.accent);
})();

(function () {
  var asked = new URLSearchParams(location.search).get("theme");
  var prefs = window.bootPrefs;
  var saved = prefs && prefs.theme;
  var theme = asked || (saved && saved !== "system" ? saved : "");
  if (theme) document.documentElement.dataset.theme = theme;
  if (prefs && prefs.colorblind) document.documentElement.dataset.colorblind = "1";
  if (prefs && prefs.reduceMotion) document.documentElement.dataset.motion = "reduce";
  if (prefs) document.documentElement.style.setProperty("--zoom", prefs.web ? 1 : (prefs.textSize || 100) / 100);
})();

// Settings' Reduce motion (#3813) applies the stylesheets' reduced-motion
// rules, their media set to all and back; calm() in app.js follows it
(function () {
  var held = [];
  window.holdStill = function (on) {
    var root = document.documentElement;
    if (on) root.dataset.motion = "reduce"; else delete root.dataset.motion;
    held.forEach(function (h) { h[0].media.mediaText = h[1]; });
    held = [];
    if (on) for (var i = 0, s = document.styleSheets; i < s.length; i++) for (var j = 0, r = s[i].cssRules; j < r.length; j++)
      if (r[j].media && /reduced-motion: *reduce/.test(r[j].media.mediaText)) held.push([r[j], r[j].media.mediaText]), r[j].media.mediaText = "all";
  };
})();
