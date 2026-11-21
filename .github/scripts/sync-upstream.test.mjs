import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';

test('upstream sync reads moepie rather than this fork', () => {
  const workflow = readFileSync(new URL('../workflows/sync-upstream.yml', import.meta.url), 'utf8');
  assert.match(workflow, /gh api repos\/lingqiqi5211\/moepie\/git\/ref\/heads\/main/);
  assert.doesNotMatch(workflow, /gh api repos\/hmjz100\/blackpie\/git\/ref\/heads\/main/);
});
