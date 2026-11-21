const assert = require("node:assert/strict");
const fs = require("node:fs/promises");
const path = require("node:path");
const { test } = require("node:test");
const { chromium, webkit } = require("playwright");

const assets = path.resolve(__dirname, "../assets");
const words = { en: "Accent color", zh: "主题色", "zh-TW": "主題色", ja: "アクセントカラー", de: "Akzentfarbe" };

function serve(settings) {
  return async (r) => {
    const url = new URL(r.request().url());
    const json = (data) => r.fulfill({ json: data });
    if (url.pathname === "/boot.js") return r.fulfill({ contentType: "text/javascript", body: `window.bootPrefs = ${JSON.stringify({ ...settings, web: true })};` });
    if (url.pathname === "/wails/runtime.js") return r.fulfill({ contentType: "text/javascript", body: "export const Window = {};" });
    if (url.pathname === "/api/state") return json({ agents: [], profiles: [], settings });
    if (url.pathname === "/api/settings") {
      if (r.request().method() === "POST") Object.assign(settings, JSON.parse(r.request().postData()));
      return json(settings);
    }
    if (url.pathname === "/api/providers") return json({ providers: [], presets: [], excluded: [], gateway: {} });
    if (url.pathname === "/api/groups") return json({ models: [], groups: [], pools: [] });
    if (url.pathname === "/api/plugins") return json({ plugins: [] });
    if (url.pathname === "/api/usage/quotas") return json([]);
    if (url.pathname.startsWith("/api/")) return json({});
    const file = path.join(assets, url.pathname === "/" ? "index.html" : url.pathname);
    const type = { ".html": "text/html", ".js": "text/javascript", ".css": "text/css", ".svg": "image/svg+xml", ".png": "image/png" }[path.extname(file)];
    await r.fulfill({ body: await fs.readFile(file), contentType: type });
  };
}

for (const engine of (process.env.BROWSER ? [process.env.BROWSER] : ["chromium", "webkit"])) {
  for (const width of [420, 1000]) for (const [lang, word] of Object.entries(words)) {
    test(`${engine} ${width} ${lang}: accent persists and remains readable`, async (t) => {
      const browser = await (engine === "webkit" ? webkit.launch() : chromium.launch({ channel: "chromium" }));
      t.after(() => browser.close());
      const page = await (await browser.newContext({ viewport: { width, height: 900 }, reducedMotion: "reduce" })).newPage();
      const settings = { lang, theme: "light", accent: "#574ab8" };
      await page.route("**/*", serve(settings));
      await page.goto("http://blackpie.test/?view=settings");
      await page.locator("#accentPick button").first().waitFor();
      assert.equal(await page.locator("#accentPick").locator("..").locator(".name").textContent(), word);
      assert.equal(await page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue("--accent").trim()), "#574ab8");
      const input = page.locator('#accentPick input[type="color"]');
      await input.fill("#0d9488");
      await page.waitForFunction(() => window.bootPrefs && getComputedStyle(document.documentElement).getPropertyValue("--accent").trim() !== "#574ab8");
      assert.equal(settings.accent, "#0d9488");
      const after = await page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue("--accent").trim());
      await page.reload();
      await page.locator("#accentPick button").first().waitFor();
      assert.equal(await page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue("--accent").trim()), after);
      assert.equal(await page.evaluate(() => document.documentElement.scrollWidth > innerWidth), false);
      await page.locator("#accentPick button").first().click();
      await page.waitForFunction(() => getComputedStyle(document.documentElement).getPropertyValue("--accent").trim() === "#574ab8");
      assert.equal(settings.accent, "");
      const readable = await page.evaluate(() => {
        const parse = (s) => [1, 3, 5].map((i) => parseInt(s.slice(i, i + 2), 16) / 255);
        const lum = (c) => c.reduce((n, v, i) => n + [0.2126, 0.7152, 0.0722][i] * (v <= .03928 ? v / 12.92 : ((v + .055) / 1.055) ** 2.4), 0);
        const ratio = (a, b) => { a = lum(parse(a)); b = lum(parse(b)); return (Math.max(a, b) + .05) / (Math.min(a, b) + .05); };
        return ["light", "dark"].every((theme) => {
          document.documentElement.dataset.theme = theme;
          window.paintAccent("#ffff00");
          const a = getComputedStyle(document.documentElement).getPropertyValue("--accent").trim();
          return ratio(a, theme === "light" ? "#f4f4f6" : "#232327") >= 4.49;
        });
      });
      assert.ok(readable, "accent must remain readable in both themes");
    });
  }
}
