#!/bin/bash

set -e

# 解析命令行参数
DRAFT_FLAG=""
VERSION=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --draft)
            DRAFT_FLAG="--draft"
            shift
            ;;
        -*)
            echo "❌ 未知选项: $1"
            echo "用法: $0 [--draft] [version]"
            exit 1
            ;;
        *)
            VERSION="$1"
            shift
            ;;
    esac
done

# 获取版本号
if [ -z "$VERSION" ]; then
    VERSION=$(git describe --tags --abbrev=0 --match='v*' 2>/dev/null)
fi

if [ -z "$VERSION" ]; then
    echo "❌ 请提供版本号: $0 [--draft] <version>"
    exit 1
fi

# 显示发布模式
if [ -n "$DRAFT_FLAG" ]; then
    echo "创建草稿版本 osbuilder $VERSION"
else
    echo "发布 osbuilder $VERSION"
fi

# 检查 gh CLI
if ! gh auth status >/dev/null 2>&1; then
    echo "❌ 请先登录: gh auth login"
    exit 1
fi

# 创建临时目录
RELEASE_DIR="_output/releases"
rm -rf "$RELEASE_DIR" && mkdir -p "$RELEASE_DIR"

# 遍历平台目录并创建压缩包
RELEASE_FILES=()
for os_dir in _output/platforms/*/; do
    os=$(basename "$os_dir")
    for arch_dir in "$os_dir"*/; do
        arch=$(basename "$arch_dir")
        binary_path="$arch_dir/osbuilder"
        
        # 检查二进制文件是否存在
        if [ ! -f "$binary_path" ]; then
            echo "⚠️  跳过不存在的文件: $binary_path"
            continue
        fi
        
        # 根据操作系统选择压缩格式
        if [ "$os" = "windows" ]; then
            # Windows 使用 zip 格式，并添加 .exe 扩展名
            archive="$RELEASE_DIR/osbuilder-${VERSION}-${os}-${arch}.zip"
            (cd "$arch_dir" && zip -q "../../../releases/$(basename "$archive")" osbuilder)
        else
            # Unix 系统使用 tar.gz 格式
            archive="$RELEASE_DIR/osbuilder-${VERSION}-${os}-${arch}.tar.gz"
            tar -czf "$archive" -C "$arch_dir" osbuilder
        fi
        
        RELEASE_FILES+=("$archive")
        echo "已创建: $(basename "$archive")"
    done
done

# 检查是否有文件要发布
if [ ${#RELEASE_FILES[@]} -eq 0 ]; then
    echo "❌ 没有找到构建产物"
    exit 1
fi

echo "共创建 ${#RELEASE_FILES[@]} 个发布文件"

# 创建标签（如果不存在且不是草稿模式）
if [ -z "$DRAFT_FLAG" ]; then
    if ! git rev-parse "$VERSION" >/dev/null 2>&1; then
        echo "创建标签 $VERSION"
        git tag -a "$VERSION" -m "Release $VERSION"
        git push origin "$VERSION"
    fi
else
    echo "草稿模式：跳过标签创建"
fi

# 创建 GitHub Release
if [ -n "$DRAFT_FLAG" ]; then
    echo "创建草稿 Release..."
    gh release create "$VERSION" \
        "${RELEASE_FILES[@]}" \
        --title "osbuilder $VERSION" \
        --generate-notes \
        --draft
    
    echo "草稿创建完成: $(gh repo view --json url -q .url)/releases/tag/$VERSION"
    echo "可以通过 GitHub 网页界面编辑并发布此草稿"
else
    echo "创建正式 Release..."
    gh release create "$VERSION" \
        "${RELEASE_FILES[@]}" \
        --title "osbuilder $VERSION" \
        --generate-notes
    
    echo "发布完成: $(gh repo view --json url -q .url)/releases/tag/$VERSION"
fi

# 清理临时文件
rm -rf "$RELEASE_DIR"
echo "清理完成"

