#!/usr/bin/env bash
# Valida la infraestructura compartida de agentes (issue #259, DEC-086).
# Solo lectura: no modifica archivos ni ejecuta herramientas externas.
#
# Uso: tools/ai/validate-agent-system.sh [--strict] [raiz-del-repo]
#   --strict  las advertencias también terminan con código distinto de cero.
set -euo pipefail

strict=0
if [[ "${1:-}" == "--strict" ]]; then
  strict=1
  shift
fi

repo_root="${1:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
failures=0
warnings=0

fail() { printf 'FAIL: %s\n' "$1"; failures=$((failures + 1)); }
warn() { printf 'WARN: %s\n' "$1"; warnings=$((warnings + 1)); }

# Valor de un campo escalar del frontmatter YAML (primer bloque --- ... ---).
frontmatter_field() {
  awk -v key="$2" '
    NR == 1 && $0 != "---" { exit }
    NR > 1 && $0 == "---" { exit }
    NR > 1 && index($0, key ": ") == 1 { print substr($0, length(key) + 3); exit }
  ' "$1"
}

rel() { printf '%s' "${1#"$repo_root/"}"; }

# --- Instrucciones globales -------------------------------------------------

for file in AGENTS.md CLAUDE.md; do
  [[ -f "$repo_root/$file" ]] || fail "missing $file"
done

if [[ -f "$repo_root/AGENTS.md" ]]; then
  agents_bytes="$(wc -c < "$repo_root/AGENTS.md")"
  (( agents_bytes <= 32768 )) || warn "AGENTS.md exceeds Codex's default 32 KiB project-doc budget ($agents_bytes bytes)"
fi

if [[ -f "$repo_root/CLAUDE.md" ]]; then
  claude_lines="$(wc -l < "$repo_root/CLAUDE.md")"
  (( claude_lines <= 200 )) || warn "CLAUDE.md exceeds Anthropic's 200-line guidance ($claude_lines lines)"
  grep -qx '@AGENTS\.md' "$repo_root/CLAUDE.md" || fail "CLAUDE.md does not import @AGENTS.md"
fi

[[ -e "$repo_root/.claude/CLAUDE.md" ]] && warn ".claude/CLAUDE.md exists; keep Claude-specific rules in CLAUDE.md only"

# --- Skills canónicos y adaptadores ----------------------------------------

shopt -s nullglob
canonical_names=()
canonical_files=("$repo_root"/.agents/skills/*/SKILL.md)
(( ${#canonical_files[@]} > 0 )) || fail "no canonical skills under .agents/skills"

for canonical in "${canonical_files[@]}"; do
  skill_path="$(dirname "$canonical")"
  skill_dir="$(basename "$skill_path")"
  canonical_names+=("$skill_dir")
  label="$(rel "$canonical")"

  [[ "$(head -n 1 "$canonical")" == '---' ]] || { fail "$label has no YAML frontmatter"; continue; }

  name="$(frontmatter_field "$canonical" name)"
  description="$(frontmatter_field "$canonical" description)"
  [[ "$name" == "$skill_dir" ]] || fail "$label name '$name' differs from directory '$skill_dir'"
  [[ "$skill_dir" =~ ^[a-z0-9]+(-[a-z0-9]+)*$ ]] || fail "$label directory is not kebab-case"
  if [[ -z "$description" ]]; then
    fail "$label has no description"
  elif [[ "$description" != \"*\" ]]; then
    fail "$label description must be a double-quoted single line"
  elif (( ${#description} > 1026 )); then
    fail "$label description exceeds 1024 characters"
  fi

  skill_lines="$(wc -l < "$canonical")"
  (( skill_lines <= 500 )) || warn "$label exceeds 500 lines; move conditional detail to references/"

  # Enlaces relativos del skill (references/, scripts/...) deben existir.
  while IFS= read -r target; do
    [[ "$target" =~ ^(https?:|mailto:|#) ]] && continue
    [[ -e "$skill_path/${target%%#*}" ]] || fail "$label links to missing file '$target'"
  done < <(grep -oE '\]\([^)]+\)' "$canonical" | sed -E 's/^\]\((.*)\)$/\1/')

  adapter="$repo_root/.claude/skills/$skill_dir/SKILL.md"
  if [[ ! -f "$adapter" ]]; then
    fail "missing Claude adapter .claude/skills/$skill_dir/SKILL.md"
    continue
  fi
  [[ "$(frontmatter_field "$adapter" name)" == "$skill_dir" ]] || fail "Claude adapter for $skill_dir has a different name"
  [[ "$(frontmatter_field "$adapter" description)" == "$description" ]] || fail "Claude adapter description for $skill_dir differs from the canonical skill"
  grep -qF ".agents/skills/$skill_dir/SKILL.md" "$adapter" || fail "Claude adapter for $skill_dir does not point to the canonical skill"
  adapter_lines="$(wc -l < "$adapter")"
  (( adapter_lines <= 20 )) || warn "Claude adapter for $skill_dir has $adapter_lines lines; content belongs in the canonical skill"
done

# Skills solo de Claude (alias heredados o herramientas) no deben competir por activación.
for claude_skill in "$repo_root"/.claude/skills/*/SKILL.md; do
  skill_dir="$(basename "$(dirname "$claude_skill")")"
  [[ " ${canonical_names[*]} " == *" $skill_dir "* ]] && continue
  [[ "$(frontmatter_field "$claude_skill" disable-model-invocation)" == "true" ]] \
    || warn "Claude-only skill $skill_dir has no canonical source and is model-invocable; share it or set disable-model-invocation: true"
done

# --- Hooks ------------------------------------------------------------------

settings="$repo_root/.claude/settings.json"
if [[ -f "$settings" ]]; then
  grep -q 'PreToolUse' "$settings" && warn ".claude/settings.json defines PreToolUse hooks; they run on every matching tool call with user permissions"
  grep -qi 'graphify' "$settings" && warn ".claude/settings.json invokes Graphify; keep it explicit (/graphify)"
  grep -qE '"[A-Za-z]:[\\/]|"/(home|Users)/' "$settings" && warn ".claude/settings.json contains a machine-specific absolute path"
fi

# graphify-out/ es regenerable y solo local (DEC-087); una reinstalación de Graphify puede volver a versionarlo.
if git -C "$repo_root" rev-parse --git-dir >/dev/null 2>&1; then
  [[ -n "$(git -C "$repo_root" ls-files -- graphify-out | head -n 1)" ]] \
    && warn "graphify-out/ is tracked by git; it must stay local (git rm -r --cached graphify-out)"
  git -C "$repo_root" check-ignore -q graphify-out/graph.json \
    || warn "graphify-out/ is not ignored in .gitignore"
fi

# --- Resultado --------------------------------------------------------------

printf '\n%d skill(s) canónico(s), %d failure(s), %d warning(s)\n' "${#canonical_names[@]}" "$failures" "$warnings"
if (( failures > 0 )) || (( strict == 1 && warnings > 0 )); then
  exit 1
fi
printf 'PASS: agent system is consistent\n'
