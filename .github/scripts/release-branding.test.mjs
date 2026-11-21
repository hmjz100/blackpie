import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';

const read = (path) => readFileSync(new URL(`../../${path}`, import.meta.url), 'utf8');

test('builds, launch checks and updater agree on the blackpie executable and bundle', () => {
  const make = read('Makefile');
  for (const name of ['-o blackpie .', 'blackpie.app/Contents/MacOS/blackpie', 'dist/blackpie-cli-', 'dist/blackpie-windows-', 'dist/blackpie-linux-']) assert.ok(make.includes(name), name);
  const workflow = read('.github/workflows/release.yml');
  for (const name of ['blackpie.app', 'blackpie-darwin-', 'dist/blackpie-cli-', 'dist\\blackpie-windows-', 'dist/blackpie-linux-', '"blackpie $env:VERSION*"', '"blackpie $VERSION"']) assert.ok(workflow.includes(name), name);
  assert.equal((read('internal/cli/main.go').match(/fmt.Println\("blackpie", version\)/g) || []).length, 2);
  assert.ok(read('internal/update/update.go').includes('filepath.Join(out, "blackpie.app")'));
  const notes = read('.github/scripts/release-notes.mjs');
  for (const name of ['blackpie-darwin-<arch>.dmg', 'blackpie-windows-<arch>.exe', 'blackpie-linux-<arch>', 'blackpie-cli-<os>-<arch>']) assert.ok(notes.includes(name), name);
});

test('visible shell branding changes without migrating stored data or import protocols', () => {
  const html = read('internal/gui/assets/index.html');
  for (const name of ['<title>blackpie</title>', '<span>blackpie</span>', 'content="blackpie"']) assert.ok(html.includes(name), name);
  const manifest = JSON.parse(read('internal/gui/assets/manifest.json'));
  assert.equal(manifest.name, 'blackpie');
  assert.equal(manifest.short_name, 'blackpie');
  assert.ok(read('internal/gui/app.go').includes('Name:           "blackpie"'));
  assert.ok(read('internal/gui/lightweight.go').includes('Title:            "blackpie"'));
  assert.ok(read('internal/gui/scheme_linux.go').includes('\nName=blackpie\n'));
  assert.ok(read('internal/gui/desktop.go').includes('const programName = "blackpie"'));
  assert.ok(read('internal/gui/scheme_windows.go').includes('Software\\Classes\\blackpie'));
  const win = JSON.parse(read('build/windows/winres.json'));
  const info = win.RT_VERSION['#1']['0000'].info['0409'];
  assert.equal(info.ProductName, 'blackpie');
  assert.equal(info.FileDescription, 'blackpie');
  assert.equal(info.OriginalFilename, 'blackpie.exe');
  assert.match(info.LegalCopyright, /hmjz100/);
  const plist = read('build/darwin/Info.plist');
  assert.match(plist, /CFBundleExecutable<\/key><string>blackpie<\/string>/);
  assert.match(plist, /io.github.hmjz100.blackpie/);
  assert.match(plist, /CFBundleURLSchemes<\/key><array><string>blackpie<\/string>/);
  for (const line of read('internal/gui/assets/i18n.js').split('\n').filter((l) => l.includes('"Open blackpie":'))) assert.ok(line.match(/blackpie/g).length === 2);
});
