#!/usr/bin/env bash
# 一键发布脚本：bump -> commit -> tag -> push -> 监控 GitHub Actions
# 用法：
#   ./scripts/release.sh patch   # x.y.z -> x.y.(z+1)
#   ./scripts/release.sh minor   # x.y.z -> x.(y+1).0
#   ./scripts/release.sh major   # x.y.z -> (x+1).0.0
#   ./scripts/release.sh v1.2.3  # 显式版本
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

# ----- 1. 计算新版本号 -----
bump_arg="${1:-patch}"
current=$(grep -E '^\s*Version\s*=\s*"' internal/config/config.go | head -1 | sed -E 's/.*"([^"]+)".*/\1/')
[ -z "$current" ] && { echo "无法从 config.go 读取当前版本"; exit 1; }

bump_semver() {
  local v=$1 part=$2
  IFS='.' read -r maj min pat <<<"${v#v}"
  case "$part" in
    major) maj=$((maj+1)); min=0; pat=0 ;;
    minor) min=$((min+1)); pat=0 ;;
    patch) pat=$((pat+1)) ;;
  esac
  echo "$maj.$min.$pat"
}

case "$bump_arg" in
  major|minor|patch) new_ver=$(bump_semver "$current" "$bump_arg") ;;
  v*)                new_ver="${bump_arg#v}" ;;
  *)                 new_ver="$bump_arg" ;;
esac
new_tag="v$new_ver"

echo "==> 当前版本: $current"
echo "==> 目标版本: $new_ver  (tag: $new_tag)"

# ----- 2. 工作区检查 -----
if ! git diff --quiet || ! git diff --cached --quiet; then
  read -rp "工作区有未提交改动，先一并发布？[y/N] " yn
  [[ "$yn" =~ ^[Yy]$ ]] || { echo "中止"; exit 1; }
fi

if git tag -l | grep -qx "$new_tag"; then
  echo "标签 $new_tag 已存在，请改用更高版本"
  exit 1
fi

# ----- 3. 同步源码版本号 -----
sed -i.bak -E "s/^(\s*Version\s*=\s*\")[^\"]+(\")/\1${new_ver}\2/" internal/config/config.go
rm -f internal/config/config.go.bak

if [ -f frontend/package.json ]; then
  sed -i.bak -E "s/(\"version\"\s*:\s*\")[^\"]+(\")/\1${new_ver}\2/" frontend/package.json
  rm -f frontend/package.json.bak
fi

if [ -f wails.json ]; then
  sed -i.bak -E "s/(\"productVersion\"\s*:\s*\")[^\"]+(\")/\1${new_ver}\2/" wails.json
  rm -f wails.json.bak
fi

# ----- 4. 生成/更新 CHANGELOG.md -----
last_tag=$(git describe --tags --abbrev=0 2>/dev/null || true)
range="HEAD"
[ -n "$last_tag" ] && range="${last_tag}..HEAD"

today=$(date +%Y-%m-%d)
notes_file=$(mktemp)
{
  echo "## $new_tag — $today"
  echo
  if [ -n "$last_tag" ]; then
    git log --no-merges --pretty='- %s' "$range" || true
  else
    git log --no-merges --pretty='- %s' || true
  fi
  echo
} > "$notes_file"

touch CHANGELOG.md
{
  echo "# Changelog"
  echo
  cat "$notes_file"
  if [ -s CHANGELOG.md ]; then
    tail -n +2 CHANGELOG.md
  fi
} > CHANGELOG.md.new && mv CHANGELOG.md.new CHANGELOG.md

# 同时把 release notes 单独存出来给 release job 使用
mkdir -p .release
cp "$notes_file" .release/notes.md
rm -f "$notes_file"

# ----- 5. 提交、打 tag、推送 -----
git add internal/config/config.go CHANGELOG.md
[ -f frontend/package.json ] && git add frontend/package.json
[ -f wails.json ] && git add wails.json

# 如果工作区还有其他改动，一并入列（前提是用户已确认）
if ! git diff --quiet; then
  git add -u
fi

git commit -m "$(cat <<EOF
release: $new_tag

$(cat .release/notes.md)
EOF
)"

git tag -a "$new_tag" -F .release/notes.md
git push origin HEAD
git push origin "$new_tag"

echo "==> 已推送 $new_tag，开始监控 GitHub Actions…"

# ----- 6. 监控 workflow -----
if ! command -v gh >/dev/null 2>&1; then
  echo "未安装 gh CLI，跳过状态监控；请到 Actions 页面查看"
  exit 0
fi

# 等到 workflow run 出现
run_id=""
for i in $(seq 1 20); do
  run_id=$(gh run list --workflow=release.yml --branch="$new_tag" --limit 1 --json databaseId --jq '.[0].databaseId' 2>/dev/null || true)
  [ -n "$run_id" ] && break
  sleep 3
done

if [ -z "$run_id" ]; then
  echo "未能定位 workflow run，请手动查看"
  exit 0
fi

echo "==> Watching run $run_id"
gh run watch "$run_id" --exit-status || {
  echo "==> Workflow 失败，可用以下命令查看失败日志："
  echo "    gh run view $run_id --log-failed"
  exit 1
}

echo "==> Release $new_tag 构建完成"
gh release view "$new_tag" --web 2>/dev/null || true
