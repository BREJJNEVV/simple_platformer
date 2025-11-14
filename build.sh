#!/bin/bash

# Cross-platform build script for Go game
# Usage: ./build.sh [clean|all|windows|linux|mac]

set -e  # Exit on any error

# Configuration
PROJECT_NAME="simple_platformer"
VERSION="1.0.0"
DIST_DIR="dist"
MAIN_FILE="main.go"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Platform targets
PLATFORMS=(
    "windows/amd64/.exe"
    "windows/386/.exe" 
    "linux/amd64/"
    "linux/386/"
    "linux/arm/"
    "linux/arm64/"
    "darwin/amd64/"
    "darwin/arm64/"
)

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_dependencies() {
    print_info "Checking dependencies..."
    
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    print_info "Go version: $(go version)"
    
    # Check if UPX is available for compression
    if command -v upx &> /dev/null; then
        UPX_AVAILABLE=1
        print_info "UPX is available for compression"
    else
        UPX_AVAILABLE=0
        print_warning "UPX not found - binaries will not be compressed"
        print_info "Install UPX: sudo apt install upx-ucl (Ubuntu) or brew install upx (macOS)"
    fi
}

clean() {
    print_info "Cleaning previous builds..."
    rm -rf "$DIST_DIR"
    go clean
    print_success "Clean completed"
}

create_dist_dir() {
    mkdir -p "$DIST_DIR"
}

build_single() {
    local os=$1
    local arch=$2
    local extension=$3
    
    local output_name="$DIST_DIR/${PROJECT_NAME}-${os}-${arch}${extension}"
    
    print_info "Building for ${os}/${arch}..."
    
    # Build command
    if GOOS="$os" GOARCH="$arch" go build -ldflags="-s -w -X main.Version=$VERSION" -o "$output_name" "$MAIN_FILE"; then
        print_success "Built: $(basename "$output_name")"
        
        # Compress with UPX if available
        if [ $UPX_AVAILABLE -eq 1 ] && [ "$os" != "windows" ] || [ "$arch" != "386" ]; then
            print_info "Compressing with UPX..."
            if upx --best "$output_name" &> /dev/null; then
                print_success "Compressed: $(basename "$output_name")"
            else
                print_warning "Compression failed for $(basename "$output_name")"
            fi
        fi
        
        # Get file size
        local size=$(du -h "$output_name" | cut -f1)
        print_info "File size: $size"
    else
        print_error "Failed to build for ${os}/${arch}"
        return 1
    fi
}

build_windows() {
    print_info "Building Windows versions..."
    build_single "windows" "amd64" ".exe"
    build_single "windows" "386" ".exe"
}

build_linux() {
    print_info "Building Linux versions..."
    build_single "linux" "amd64" ""
    build_single "linux" "386" ""
    build_single "linux" "arm" ""
    build_single "linux" "arm64" ""
}

build_mac() {
    print_info "Building macOS versions..."
    build_single "darwin" "amd64" ""
    build_single "darwin" "arm64" ""
}

build_all() {
    print_info "Starting cross-platform build for $PROJECT_NAME v$VERSION"
    
    check_dependencies
    create_dist_dir
    
    local total=0
    local successful=0
    
    for platform in "${PLATFORMS[@]}"; do
        IFS='/' read -r os arch extension <<< "$platform"
        if build_single "$os" "$arch" "$extension"; then
            ((successful++))
        fi
        ((total++))
    done
    
    print_success "Build completed: $successful/$total platforms successful"
}

create_archive() {
    print_info "Creating distribution archives..."
    
    cd "$DIST_DIR"
    
    for file in *; do
        if [ -f "$file" ] && [ -x "$file" ] || [[ "$file" == *.exe ]]; then
            local base_name="${file%.*}"
            local extension="${file##*.}"
            
            if [[ "$file" == *"windows"* ]]; then
                # For Windows, create zip with supporting files
                mkdir -p "$base_name"
                cp "$file" "$base_name/"
                # Copy additional files if they exist
                [ -f ../README.md ] && cp ../README.md "$base_name/"
                [ -f ../LICENSE ] && cp ../LICENSE "$base_name/"
                zip -r "${base_name}.zip" "$base_name" > /dev/null
                rm -rf "$base_name"
                print_success "Created: ${base_name}.zip"
            else
                # For Unix-like, create tar.gz
                tar czf "${base_name}.tar.gz" "$file" > /dev/null
                print_success "Created: ${base_name}.tar.gz"
            fi
        fi
    done
    
    cd ..
}

show_summary() {
    print_info "=== Build Summary ==="
    echo "Project: $PROJECT_NAME v$VERSION"
    echo "Output directory: $DIST_DIR/"
    echo ""
    
    if [ -d "$DIST_DIR" ]; then
        echo "Generated files:"
        for file in "$DIST_DIR"/*; do
            if [ -f "$file" ]; then
                local size=$(du -h "$file" | cut -f1)
                printf "  %-40s %s\n" "$(basename "$file")" "$size"
            fi
        done
    fi
}

# Main script logic
case "${1:-all}" in
    "clean")
        clean
        ;;
    "windows")
        check_dependencies
        create_dist_dir
        build_windows
        ;;
    "linux")
        check_dependencies
        create_dist_dir
        build_linux
        ;;
    "mac"|"darwin")
        check_dependencies
        create_dist_dir
        build_mac
        ;;
    "all")
        clean
        build_all
        create_archive
        ;;
    "summary")
        show_summary
        ;;
    *)
        echo "Usage: $0 [clean|all|windows|linux|mac|summary]"
        echo ""
        echo "Commands:"
        echo "  clean     - Remove build artifacts"
        echo "  all       - Build for all platforms (default)"
        echo "  windows   - Build only Windows versions"
        echo "  linux     - Build only Linux versions"
        echo "  mac       - Build only macOS versions"
        echo "  summary   - Show build summary"
        exit 1
        ;;
esac

show_summary