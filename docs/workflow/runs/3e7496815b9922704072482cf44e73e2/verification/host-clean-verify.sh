#!/bin/bash
set -euo pipefail

label=${1:?Supply a unique host execution label}
[[ $label =~ ^[a-zA-Z0-9-]+$ ]] || exit 2
source_root=$(git rev-parse --show-toplevel)
cd "$source_root"
evidence_root=docs/workflow/runs/3e7496815b9922704072482cf44e73e2/verification
prefix="$source_root/$evidence_root/host-clean-$label"
[[ ! -e "$prefix.log" && ! -e "$prefix.state.log" ]] || exit 2
browser_outputs="$source_root/.cache/host-clean-$label-test-results"
[[ ! -e "$browser_outputs" ]] || exit 2
temporary=$(mktemp -d "${TMPDIR:-/tmp}/model-relay-clean.XXXXXXXX")
trap 'rm -rf "$temporary"' EXIT
mkdir "$temporary/candidate"
git ls-files --cached --others --exclude-standard -z > "$temporary/files"
tar --null -T "$temporary/files" -cf "$temporary/source.tar"
tar -xf "$temporary/source.tar" -C "$temporary/candidate"
export GIT_DIR=$(git rev-parse --absolute-git-dir)
export GIT_WORK_TREE="$temporary/candidate"
unset NODE_PATH NODE_OPTIONS
cd "$temporary/candidate"
export GOCACHE="$temporary/candidate/.cache/go-build"
export GOTOOLCHAIN=local
export CGO_ENABLED=1
test ! -e node_modules
test ! -e .cache/npm
go run ./tests/evidence > "$temporary/source-before.json"
grep -Fq '"working_tree_sha256":"8e82aaa123df25a4758a1939fb2d018713831507aee337e18dfa6100c49f619e"' "$temporary/source-before.json"
{
  printf 'candidate=isolated-copy-%s\ngit_history=original-read-only\nnode_modules_before=absent\nlocal_npm_cache_before=absent\nNODE_PATH=unset\nNODE_OPTIONS=unset\nmanual_npm_preinstall=none\ncommand=make verify\n' "$label"
  printf 'started_utc=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  cat "$temporary/source-before.json"
  go version
  node --version
  npm --version
} > "$temporary/state"
set +e
make verify > "$temporary/verify.log" 2>&1
verify_exit=$?
set -e
{
  printf 'finished_utc=%s\nmake_verify_exit=%s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$verify_exit"
  go run ./tests/evidence > "$temporary/source-after.json" || verify_exit=1
  cat "$temporary/source-after.json"
  if [[ -d node_modules ]]; then
    node -e 'const path=require("node:path"); for (const name of ["playwright", "playwright-core"]) { const resolved=require.resolve(name+"/package.json"); if (!resolved.startsWith(path.join(process.cwd(),"node_modules")+path.sep)) process.exit(1); const version=require(resolved).version; console.log(name+"="+version+"; resolution=candidate/node_modules"); if(version!=="1.59.1") process.exit(1); }' || verify_exit=1
    node -e 'const {chromium}=require("playwright"); const result=require("node:child_process").spawnSync(chromium.executablePath(),["--version"],{encoding:"utf8"}); if(result.status!==0) { console.log("chromium_version_capture=failed"); process.exit(1); } console.log("chromium="+result.stdout.trim());' || verify_exit=1
  elif [[ $verify_exit -eq 0 ]]; then
    verify_exit=1
  fi
  cmp -s "$temporary/source-before.json" "$temporary/source-after.json" || verify_exit=1
  printf 'capture_exit=%s\n' "$verify_exit"
} >> "$temporary/state"
redact() {
  while IFS= read -r line || [[ -n $line ]]; do
    line=${line//"$source_root"/<source>}
    line=${line//"$temporary"/<isolated>}
    line=${line//"${HOME:-__no_home__}"/<home>}
    printf '%s\n' "$line"
  done
}
redact < "$temporary/verify.log" > "$prefix.log"
redact < "$temporary/state" > "$prefix.state.log"
shasum -a 256 "$prefix.log" "$prefix.state.log" | redact > "$prefix.sha256"
if [[ -d test-results ]]; then
  mkdir -p "$source_root/.cache"
  cp -R test-results "$browser_outputs"
fi
printf 'receipt=%s/host-clean-%s.state.log\nfull_log=%s/host-clean-%s.log\nexit_code=%s\n' "$evidence_root" "$label" "$evidence_root" "$label" "$verify_exit"
exit "$verify_exit"
