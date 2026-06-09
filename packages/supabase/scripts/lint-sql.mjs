// Dependency-free format lint for the project's SQL (migrations + RLS tests).
//
// We don't want to drag a heavy SQL toolchain (sqlfluff/pgFormatter) into the
// monorepo just to keep whitespace honest, so this enforces the same baseline
// hygiene Prettier gives the TypeScript: LF line endings, a single trailing
// newline, no trailing whitespace, and spaces-not-tabs for indentation. It
// runs over every *.sql file under migrations/ and tests/ and exits non-zero
// with a precise file:line report on the first batch of violations.

import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join, relative } from 'node:path';
import { fileURLToPath } from 'node:url';

const packageRoot = fileURLToPath(new URL('..', import.meta.url));
const SQL_DIRS = ['migrations', 'tests'];

function sqlFilesIn(dir) {
  const abs = join(packageRoot, dir);
  let entries;
  try {
    entries = readdirSync(abs);
  } catch {
    return []; // Directory may not exist yet — nothing to lint.
  }
  return entries
    .filter((name) => name.endsWith('.sql'))
    .map((name) => join(abs, name))
    .filter((p) => statSync(p).isFile());
}

function lintFile(absPath) {
  const rel = relative(packageRoot, absPath);
  const raw = readFileSync(absPath, 'utf8');
  const problems = [];

  if (raw.includes('\r')) {
    problems.push(`${rel}: uses CRLF line endings (expected LF)`);
  }
  if (raw.length > 0 && !raw.endsWith('\n')) {
    problems.push(`${rel}: missing trailing newline`);
  }
  if (raw.endsWith('\n\n')) {
    problems.push(`${rel}: multiple trailing newlines`);
  }

  const lines = raw.split('\n');
  lines.forEach((line, i) => {
    const lineNo = i + 1;
    if (/[ \t]+$/.test(line)) {
      problems.push(`${rel}:${lineNo}: trailing whitespace`);
    }
    if (/^\t| \t/.test(line)) {
      problems.push(`${rel}:${lineNo}: tab indentation (use spaces)`);
    }
  });

  return problems;
}

const files = SQL_DIRS.flatMap(sqlFilesIn);
const problems = files.flatMap(lintFile);

if (problems.length > 0) {
  console.error('SQL format lint failed:');
  for (const problem of problems) console.error(`  ${problem}`);
  process.exit(1);
}

console.log(`SQL format lint passed (${files.length} file(s) checked).`);
