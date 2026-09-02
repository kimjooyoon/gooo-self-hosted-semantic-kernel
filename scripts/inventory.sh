#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
output="${1:-$repo_root/evidence/inventory.json}"

cd "$repo_root"

count_files() {
	find . -type f -not -path './.git/*' -not -path './README.md' -print | wc -l | tr -d ' '
}

count_dirs() {
	find . -type d -not -path './.git' -not -path './.git/*' -print | wc -l | tr -d ' '
}

count_named_files() {
	local pattern="$1"
	find . -type f -name "$pattern" -not -path './.git/*' -print | wc -l | tr -d ' '
}

line_count() {
	local pattern="$1"
	local count
	count="$(find . -type f -name "$pattern" -not -path './.git/*' -exec awk 'END { total += NR } END { print total + 0 }' {} +)"
	printf '%s' "${count:-0}"
}

generated_count="$(find generated -type f -not -name '.gitkeep' -not -path 'generated/.git/*' -print 2>/dev/null | wc -l | tr -d ' ')"

jq -n \
	--argjson directories "$(count_dirs)" \
	--argjson files "$(count_files)" \
	--argjson go_files "$(count_named_files '*.go')" \
	--argjson go_lines "$(line_count '*.go')" \
	--argjson gooo_files "$(count_named_files '*.gooo')" \
	--argjson gooo_lines "$(line_count '*.gooo')" \
	--argjson generated_files "$generated_count" \
	--argjson wall_ms "${GOOO_WALL_MS:-null}" \
	--argjson peak_rss_kib "${GOOO_PEAK_RSS_KIB:-null}" \
	'{schema:"gooo.inventory/v1", root_readme_excluded:true, directories:$directories, files:$files, go_files:$go_files, go_lines:$go_lines, gooo_files:$gooo_files, gooo_lines:$gooo_lines, generated_files:$generated_files, wall_ms:$wall_ms, peak_rss_kib:$peak_rss_kib}' > "$output"
