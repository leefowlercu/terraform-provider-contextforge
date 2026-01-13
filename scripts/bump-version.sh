#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Validate arguments
if [ $# -ne 1 ]; then
    echo -e "${RED}Error: Bump type required${NC}"
    echo "Usage: $0 <major|minor|patch>"
    exit 1
fi

BUMP_TYPE=$1

if [[ ! "$BUMP_TYPE" =~ ^(major|minor|patch)$ ]]; then
    echo -e "${RED}Error: Invalid bump type '${BUMP_TYPE}'${NC}"
    echo "Must be one of: major, minor, patch"
    exit 1
fi

# Get current version from git tags (most recent semver tag)
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

# Strip 'v' prefix if present
CURRENT_VERSION="${LATEST_TAG#v}"

if [ -z "$CURRENT_VERSION" ] || [ "$CURRENT_VERSION" = "0.0.0" ]; then
    echo -e "${YELLOW}Warning: No existing version tags found, starting from 0.0.0${NC}"
    CURRENT_VERSION="0.0.0"
fi

echo -e "${GREEN}Current version: ${CURRENT_VERSION}${NC}"

# Validate current version is valid semver (X.Y.Z)
if [[ ! "$CURRENT_VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}Error: Invalid version format '${CURRENT_VERSION}'${NC}"
    echo "Expected format: X.Y.Z (e.g., 0.1.0)"
    exit 1
fi

# Parse version components
IFS='.' read -r MAJOR MINOR PATCH <<< "$CURRENT_VERSION"

# Calculate new version based on bump type
case "$BUMP_TYPE" in
    major)
        MAJOR=$((MAJOR + 1))
        MINOR=0
        PATCH=0
        ;;
    minor)
        MINOR=$((MINOR + 1))
        PATCH=0
        ;;
    patch)
        PATCH=$((PATCH + 1))
        ;;
esac

NEW_VERSION="${MAJOR}.${MINOR}.${PATCH}"
echo -e "${GREEN}New version: ${NEW_VERSION}${NC}"

# Write new version with 'v' prefix to temp file for Makefile to read
echo "v${NEW_VERSION}" > "${PROJECT_ROOT}/.next-version"
echo -e "${GREEN}Wrote v${NEW_VERSION} to .next-version${NC}"
