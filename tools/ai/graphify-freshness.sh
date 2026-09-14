#!/usr/bin/env bash
# Decide si el grafo local de Graphify necesita actualizarse (issue #264, DEC-088).
# Solo lectura: no ejecuta Graphify ni modifica archivos.
#
# Uso: tools/ai/graphify-freshness.sh [raiz-del-repo]
#   GRAPHIFY_DOCS_THRESHOLD  archivos documentales cambiados desde la última
#                            reextracción semántica para sugerirla (defecto 30).
#
# Salida: líneas clave=valor y, al final, `recommendation=`:
#   BUILD_FIRST             no existe graphify-out/graph.json.
#   UPDATE                  se integró código desde built_at_commit: `graphify update .`.
#   SEMANTIC_UPDATE_SUGGESTED  bloque documental grande desde la última reextracción
#                           semántica: proponer `/graphify . --update` (usa IA, pedir confirmación).
#   SKIP                    el grafo sirve tal como está.
set -euo pipefail

repo_root="${1:-$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)}"
docs_threshold="${GRAPHIFY_DOCS_THRESHOLD:-30}"
out_dir="$repo_root/graphify-out"
graph="$out_dir/graph.json"
semantic_commit_file="$out_dir/.graphify_semantic_commit"
semantic_marker="$out_dir/.graphify_semantic_marker"

g() { git -C "$repo_root" "$@"; }

g rev-parse --git-dir >/dev/null 2>&1 || { echo "error=not a git repository: $repo_root" >&2; exit 2; }
head_sha="$(g rev-parse HEAD)"

# Hooks git de Graphify: DEC-087 los retiró; una reinstalación puede reponerlos.
hooks="none"
hooks_dir="$(g rev-parse --path-format=absolute --git-path hooks)"
for hook in post-commit post-checkout; do
  if [[ -f "$hooks_dir/$hook" ]] && grep -q 'graphify' "$hooks_dir/$hook"; then
    hooks="installed"
  fi
done
echo "hooks=$hooks"
[[ "$hooks" == "installed" ]] && echo "warning=Graphify git hooks are installed again; remove them with: graphify hook uninstall"

if [[ ! -f "$graph" ]]; then
  echo "graph=missing"
  echo "recommendation=BUILD_FIRST"
  exit 0
fi
echo "graph=present"

# Clasifica rutas cambiadas: la actualización normal solo relee código; la
# documentación solo entra en la reextracción semántica (con IA).
classify() {
  local code=0 docs=0 path
  while IFS= read -r path; do
    [[ -z "$path" ]] && continue
    case "$path" in
      graphify-out/*|apps/web/e2e/evidence/*|*.png|*.jpg|*.jpeg|*.webp|*.gif|*.svg|*.ico|pnpm-lock.yaml|*/pnpm-lock.yaml|*.sum)
        ;;
      docs/*|respuesta-manuales/*|*.md|*.mdx|*.txt|*.docx|*.pdf)
        docs=$((docs + 1)) ;;
      *)
        code=$((code + 1)) ;;
    esac
  done
  echo "$code $docs"
}

changed_since() {
  if g cat-file -e "$1^{commit}" 2>/dev/null; then
    g diff --name-only "$1" "$head_sha"
  else
    return 1
  fi
}

built_at="$(grep -o '"built_at_commit": *"[0-9a-f]\{7,40\}"' "$graph" | tail -n 1 | grep -o '[0-9a-f]\{7,40\}' || true)"
code_changes=-1
if [[ -n "$built_at" ]] && changes="$(changed_since "$built_at")"; then
  read -r code_changes _ < <(classify <<<"$changes")
  echo "built_at_commit=$built_at"
  echo "commits_behind=$(g rev-list --count "$built_at..$head_sha" 2>/dev/null || echo unknown)"
else
  echo "built_at_commit=${built_at:-unknown}"
  echo "note=built_at_commit is missing or not in this clone; treating the graph as stale"
fi
echo "code_changes=$code_changes"

# Base de la última reextracción semántica: marcador propio escrito por el
# skill o, si no existe, el commit vigente cuando Graphify escribió su marcador.
semantic_base=""
if [[ -f "$semantic_commit_file" ]]; then
  semantic_base="$(tr -d '[:space:]' < "$semantic_commit_file")"
elif [[ -f "$semantic_marker" ]]; then
  semantic_base="$(g rev-list -1 --before="@$(stat -c %Y "$semantic_marker")" "$head_sha" 2>/dev/null || true)"
fi
docs_changes=-1
if [[ -n "$semantic_base" ]] && changes="$(changed_since "$semantic_base")"; then
  read -r _ docs_changes < <(classify <<<"$changes")
  echo "semantic_base_commit=$semantic_base"
else
  echo "semantic_base_commit=unknown"
fi
echo "docs_changes_since_semantic_update=$docs_changes"
echo "docs_threshold=$docs_threshold"

if (( docs_changes >= docs_threshold )); then
  echo "recommendation=SEMANTIC_UPDATE_SUGGESTED"
elif (( code_changes != 0 )); then
  echo "recommendation=UPDATE"
else
  echo "recommendation=SKIP"
fi
