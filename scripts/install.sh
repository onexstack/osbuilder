
#!/bin/bash

set -e

# 默认配置
REPO="onexstack/osbuilder"
BINARY_NAME="osbuilder"
INSTALL_DIR="./"
VERSION="${VERSION:-latest}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# 显示帮助信息
show_help() {
    cat << 'EOF'
osbuilder 安装脚本

用法:
  curl -fsSL https://raw.githubusercontent.com/onexstack/osbuilder/main/install.sh | bash
  curl -fsSL https://raw.githubusercontent.com/onexstack/osbuilder/main/install.sh | bash -s -- [选项]

选项:
  -v, --version VERSION    指定版本 (默认: latest)
  -d, --dir DIRECTORY      安装目录 (默认: /usr/local/bin)
  -h, --help              显示帮助信息

环境变量:
  VERSION                  指定版本号
  INSTALL_DIR              安装目录

示例:
  # 安装最新版本
  curl -fsSL https://example.com/install.sh | bash
  
  # 安装指定版本
  curl -fsSL https://example.com/install.sh | bash -s -- --version v0.8.0
  
  # 安装到用户目录
  curl -fsSL https://example.com/install.sh | bash -s -- --dir ~/.local/bin

EOF
}

# 解析命令行参数
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--version)
                VERSION="$2"
                shift 2
                ;;
            -d|--dir)
                INSTALL_DIR="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "未知选项: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# 检测操作系统
detect_os() {
    case "$(uname -s)" in
        Linux*)     OS="linux" ;;
        Darwin*)    OS="darwin" ;;
        CYGWIN*|MINGW*|MSYS*) OS="windows" ;;
        *)
            log_error "不支持的操作系统: $(uname -s)"
            exit 1
            ;;
    esac
    log_info "检测到操作系统: $OS"
}

# 检测架构
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   ARCH="amd64" ;;
        aarch64|arm64)  ARCH="arm64" ;;
        armv7l|armv6l)  ARCH="arm" ;;
        i386|i686)      ARCH="386" ;;
        *)
            log_error "不支持的架构: $(uname -m)"
            exit 1
            ;;
    esac
    log_info "检测到架构: $ARCH"
}

# 获取最新版本
get_latest_version() {
    if [[ "$VERSION" == "latest" ]]; then
        log_info "获取最新版本信息..."
        VERSION=$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
        if [[ -z "$VERSION" ]]; then
            log_error "无法获取最新版本信息"
            exit 1
        fi
    fi
    log_info "目标版本: $VERSION"
}

# 构建下载 URL
build_download_url() {
    # 确定文件扩展名
    if [[ "$OS" == "windows" ]]; then
        ARCHIVE_EXT="zip"
        BINARY_EXT=".exe"
    else
        ARCHIVE_EXT="tar.gz"
        BINARY_EXT=""
    fi
    
    ARCHIVE_NAME="${BINARY_NAME}-${VERSION}-${OS}-${ARCH}.${ARCHIVE_EXT}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE_NAME}"
    
    log_info "下载 URL: $DOWNLOAD_URL"
}

# 检查工具依赖
check_dependencies() {
    local missing_tools=()
    
    # 必需工具
    for tool in curl; do
        if ! command -v "$tool" >/dev/null 2>&1; then
            missing_tools+=("$tool")
        fi
    done
    
    # 根据文件类型检查解压工具
    if [[ "$ARCHIVE_EXT" == "tar.gz" ]]; then
        if ! command -v tar >/dev/null 2>&1; then
            missing_tools+=("tar")
        fi
    elif [[ "$ARCHIVE_EXT" == "zip" ]]; then
        if ! command -v unzip >/dev/null 2>&1; then
            missing_tools+=("unzip")
        fi
    fi
    
    if [[ ${#missing_tools[@]} -gt 0 ]]; then
        log_error "缺少必需工具: ${missing_tools[*]}"
        log_info "请安装缺少的工具后重试"
        exit 1
    fi
}

# 创建临时目录
create_temp_dir() {
    TEMP_DIR=$(mktemp -d)
    log_info "创建临时目录: $TEMP_DIR"
    
    # 设置清理陷阱
    trap 'rm -rf "$TEMP_DIR"' EXIT
}

# 下载文件
download_file() {
    log_info "下载 $ARCHIVE_NAME..."
    
    local archive_path="$TEMP_DIR/$ARCHIVE_NAME"
    
    if ! curl -fsSL -o "$archive_path" "$DOWNLOAD_URL"; then
        log_error "下载失败: $DOWNLOAD_URL"
        log_info "请检查版本号是否正确，或访问 https://github.com/$REPO/releases 查看可用版本"
        exit 1
    fi
    
    log_success "下载完成"
}

# 解压文件
extract_file() {
    log_info "解压文件..."
    
    local archive_path="$TEMP_DIR/$ARCHIVE_NAME"
    
    if [[ "$ARCHIVE_EXT" == "tar.gz" ]]; then
        tar -xzf "$archive_path" -C "$TEMP_DIR"
    elif [[ "$ARCHIVE_EXT" == "zip" ]]; then
        unzip -q "$archive_path" -d "$TEMP_DIR"
    fi
    
    log_success "解压完成"
}

# 检查安装目录权限
check_install_permissions() {
    if [[ ! -d "$INSTALL_DIR" ]]; then
        log_warn "安装目录不存在，尝试创建: $INSTALL_DIR"
        if ! mkdir -p "$INSTALL_DIR" 2>/dev/null; then
            log_error "无法创建安装目录: $INSTALL_DIR"
            log_info "请使用 sudo 运行此脚本，或选择有写权限的目录"
            log_info "例如: curl -fsSL ... | bash -s -- --dir ~/.local/bin"
            exit 1
        fi
    fi
    
    if [[ ! -w "$INSTALL_DIR" ]]; then
        log_warn "需要管理员权限安装到: $INSTALL_DIR"
        USE_SUDO=true
    else
        USE_SUDO=false
    fi
}

# 安装二进制文件
install_binary() {
    log_info "安装到 $INSTALL_DIR..."
    
    local binary_path="$TEMP_DIR/${BINARY_NAME}${BINARY_EXT}"
    local install_path="$INSTALL_DIR/${BINARY_NAME}${BINARY_EXT}"
    
    # 检查二进制文件是否存在
    if [[ ! -f "$binary_path" ]]; then
        log_error "找不到二进制文件: $binary_path"
        log_info "压缩包内容:"
        ls -la "$TEMP_DIR/"
        exit 1
    fi
    
    # 复制文件
    if [[ "$USE_SUDO" == "true" ]]; then
        sudo cp "$binary_path" "$install_path"
        sudo chmod +x "$install_path"
    else
        cp "$binary_path" "$install_path"
        chmod +x "$install_path"
    fi
    
    log_success "安装完成: $install_path"
}

# 验证安装
verify_installation() {
    log_info "验证安装..."
    
    local binary_path="$INSTALL_DIR/${BINARY_NAME}${BINARY_EXT}"
    
    if [[ ! -x "$binary_path" ]]; then
        log_error "安装验证失败: $binary_path 不可执行"
        exit 1
    fi
    
    # 检查版本
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local installed_version
        installed_version=$("$BINARY_NAME" version 2>/dev/null | head -1 || echo "unknown")
        log_success "安装版本: $installed_version"
    else
        log_warn "$INSTALL_DIR 不在 PATH 中"
        log_info "请将以下行添加到 ~/.bashrc 或 ~/.zshrc:"
        log_info "export PATH=\"$INSTALL_DIR:\$PATH\""
    fi
}

# 显示完成信息
show_completion() {
    echo
    log_success "osbuilder 安装成功!"
    echo
    log_info "使用方法:"
    echo -e "  ${BINARY_NAME} --help"
    echo -e "  ${BINARY_NAME} version"
    echo
    
    if ! command -v "$BINARY_NAME" >/dev/null 2>&1; then
        log_info "如果命令未找到，请运行:"
        echo -e "  export PATH=\"$INSTALL_DIR:\$PATH\""
        echo
    fi
}

# 主函数
main() {
    echo -e "${BLUE}osbuilder 安装脚本${NC}"
    echo
    
    # 解析参数
    parse_args "$@"
    
    # 检测系统信息
    detect_os
    detect_arch
    
    # 获取版本信息
    get_latest_version
    
    # 构建下载 URL
    build_download_url
    
    # 检查依赖
    check_dependencies
    
    # 检查安装权限
    check_install_permissions
    
    # 创建临时目录
    create_temp_dir
    
    # 下载和解压
    download_file
    extract_file
    
    # 安装
    install_binary
    
    # 验证
    verify_installation
    
    # 显示完成信息
    show_completion
}

# 运行主函数
main "$@"
