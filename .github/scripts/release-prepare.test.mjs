import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { test } from 'node:test';
import { runInNewContext } from 'node:vm';

const source = readFileSync(new URL('./release-prepare.mjs', import.meta.url), 'utf8').replace(/^import .*;\n/gm, '');
const sha = 'a'.repeat(40);
const repo = 'hmjz100/blackpie';
const valid = { head_sha: sha, head_branch: 'main', head_repository: { full_name: repo }, conclusion: 'success' };

function prepare(runs) {
  const outputs = [];
  const execFileSync = (cmd, args) => {
    if (cmd === 'git') {
      if (args[0] === 'rev-parse') return sha;
      return '';
    }
    if (args[1].includes('/actions/workflows/')) {
      const event = new URLSearchParams(args[1].split('?')[1]).get('event');
      return JSON.stringify({ workflow_runs: runs[event] || [] });
    }
    if (args[1].includes('/releases?')) return '[]';
    throw new Error(`unexpected command: ${cmd} ${args.join(' ')}`);
  };
  runInNewContext(source, {
    execFileSync, appendFileSync: (file, value) => outputs.push([file, value]),
    process: { env: { GITHUB_REPOSITORY: repo, GITHUB_OUTPUT: 'output', GITHUB_STEP_SUMMARY: 'summary' } },
  });
  return outputs;
}

for (const event of ['push', 'workflow_dispatch']) {
  test(`release accepts a successful ${event} Test on the exact main commit`, () => {
    assert.ok(prepare({ [event]: [{ ...valid, event }] }).some(([file, value]) => file === 'output' && value.includes(`sha=${sha}`)));
  });
  for (const change of [
    { head_sha: 'b'.repeat(40) }, { head_branch: 'other' },
    { head_repository: { full_name: 'other/blackpie' } }, { conclusion: 'failure' },
    { event: 'pull_request' },
  ]) {
    test(`release rejects ${event} Test with ${JSON.stringify(change)}`, () => {
      assert.throws(() => prepare({ [event]: [{ ...valid, event, ...change }] }), /Test has not passed/);
    });
  }
}

test('release rejects a commit with no successful Test', () => {
  assert.throws(() => prepare({}), /Test has not passed/);
});
