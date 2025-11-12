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
VERSION_FILE="${PROJECT_ROOT}/main.go"
CHANGELOG="${PROJECT_ROOT}/CHANGELOG.md"
DIST_CHANGELOG="${PROJECT_ROOT}/dist/CHANGELOG.md"

# Cleanup function to remove temp files on exit
cleanup() {
    rm -f "${PROJECT_ROOT}/.next-version"
}
trap cleanup EXIT

# Function to merge goreleaser changelog into main CHANGELOG.md
merge_changelog() {
    local VERSION=$1

    echo -e "${YELLOW}Merging goreleaser changelog into CHANGELOG.md...${NC}"

    # Check if dist/CHANGELOG.md exists
    if [ ! -f "$DIST_CHANGELOG" ]; then
        echo -e "${YELLOW}Warning: ${DIST_CHANGELOG} not found, skipping changelog merge${NC}"
        return 0
    fi

    # Check if dist/CHANGELOG.md is empty
    if [ ! -s "$DIST_CHANGELOG" ]; then
        echo -e "${YELLOW}Warning: ${DIST_CHANGELOG} is empty, skipping changelog merge${NC}"
        return 0
    fi

    # Check if version already exists in changelog
    if grep -q "## \[${VERSION}\]" "$CHANGELOG"; then
        echo -e "${RED}Error: Version ${VERSION} already exists in CHANGELOG.md${NC}"
        return 1
    fi

    # Parse goreleaser changelog and transform to Keep a Changelog format
    local TEMP_CHANGELOG=$(mktemp)
    local IN_SECTION=false
    local SECTION_NAME=""
    local ENTRIES=""

    # Read goreleaser changelog and extract sections
    while IFS= read -r line; do
        # Check for section headers (e.g., "### Added", "### Fixed")
        if [[ "$line" =~ ^###[[:space:]]+(.*) ]]; then
            # Save previous section if it had entries
            if [ -n "$ENTRIES" ] && [ -n "$SECTION_NAME" ]; then
                echo -e "\n### $SECTION_NAME\n" >> "$TEMP_CHANGELOG"
                echo "$ENTRIES" >> "$TEMP_CHANGELOG"
            fi

            SECTION_NAME="${BASH_REMATCH[1]}"
            ENTRIES=""
            IN_SECTION=true
        # Check for commit entries (e.g., "* abc1234 commit message")
        elif [[ "$line" =~ ^\*[[:space:]]+[a-f0-9]{7}[[:space:]]+(.*) ]]; then
            local MESSAGE="${BASH_REMATCH[1]}"
            # Capitalize first letter and format as list item
            MESSAGE="$(echo "$MESSAGE" | sed 's/^./\U&/')"
            ENTRIES="${ENTRIES}- ${MESSAGE}\n"
        fi
    done < "$DIST_CHANGELOG"

    # Save last section
    if [ -n "$ENTRIES" ] && [ -n "$SECTION_NAME" ]; then
        echo -e "\n### $SECTION_NAME\n" >> "$TEMP_CHANGELOG"
        echo -e "$ENTRIES" >> "$TEMP_CHANGELOG"
    fi

    # Check if we extracted any changelog entries
    if [ ! -s "$TEMP_CHANGELOG" ]; then
        echo -e "${YELLOW}Warning: No changelog entries extracted, skipping merge${NC}"
        rm -f "$TEMP_CHANGELOG"
        return 0
    fi

    # Get date from git tag
    local TAG_DATE=$(git log -1 --format=%ai "$VERSION" | cut -d' ' -f1)

    # Build new version section
    local VERSION_SECTION="## [${VERSION}] - ${TAG_DATE}\n"
    VERSION_SECTION="${VERSION_SECTION}$(cat "$TEMP_CHANGELOG")"

    # Find insertion point (after header, before first version or at end)
    local INSERT_LINE=$(grep -n "^## \[" "$CHANGELOG" | head -1 | cut -d: -f1)

    if [ -z "$INSERT_LINE" ]; then
        # No existing versions, append before links section if it exists
        INSERT_LINE=$(grep -n "^\[" "$CHANGELOG" | head -1 | cut -d: -f1)
        if [ -n "$INSERT_LINE" ]; then
            INSERT_LINE=$((INSERT_LINE - 1))
        fi
    fi

    if [ -n "$INSERT_LINE" ]; then
        # Insert at specific line
        {
            head -n $((INSERT_LINE - 1)) "$CHANGELOG"
            echo -e "$VERSION_SECTION"
            tail -n +$INSERT_LINE "$CHANGELOG"
        } > "${CHANGELOG}.tmp"
    else
        # Append to end
        {
            cat "$CHANGELOG"
            echo -e "\n$VERSION_SECTION"
        } > "${CHANGELOG}.tmp"
    fi

    mv "${CHANGELOG}.tmp" "$CHANGELOG"

    # Update comparison links
    local PREVIOUS_VERSION=$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")

    # Check if links section exists
    if ! grep -q "^\[${VERSION}\]:" "$CHANGELOG"; then
        if [ -n "$PREVIOUS_VERSION" ]; then
            # Add comparison link
            echo "" >> "$CHANGELOG"
            echo "[${VERSION}]: https://github.com/leefowlercu/terraform-provider-contextforge/compare/${PREVIOUS_VERSION}...${VERSION}" >> "$CHANGELOG"
        else
            # First release, link to tag
            echo "" >> "$CHANGELOG"
            echo "[${VERSION}]: https://github.com/leefowlercu/terraform-provider-contextforge/releases/tag/${VERSION}" >> "$CHANGELOG"
        fi
    fi

    rm -f "$TEMP_CHANGELOG"
    echo -e "${GREEN}Successfully merged changelog${NC}"

    return 0
}

# Validate arguments
if [ $# -ne 1 ]; then
    echo -e "${RED}Error: Version required${NC}"
    echo "Usage: $0 <vX.Y.Z>"
    exit 1
fi

VERSION=$1

# Validate version format (vX.Y.Z)
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}Error: Invalid version format '${VERSION}'${NC}"
    echo "Expected format: vX.Y.Z (e.g., v0.1.0)"
    exit 1
fi

# Remove 'v' prefix for version file
VERSION_NO_V="${VERSION#v}"

echo -e "${GREEN}Preparing release ${VERSION}...${NC}"

# Check if tag already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    echo -e "${RED}Error: Tag ${VERSION} already exists${NC}"
    exit 1
fi

# Update version in main.go
echo -e "${YELLOW}Updating version in main.go...${NC}"
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS sed requires explicit backup extension
    sed -i '' "s/version = \"[0-9]\+\.[0-9]\+\.[0-9]\+\"/version = \"${VERSION_NO_V}\"/" "$VERSION_FILE"
else
    sed -i "s/version = \"[0-9]\+\.[0-9]\+\.[0-9]\+\"/version = \"${VERSION_NO_V}\"/" "$VERSION_FILE"
fi

# Stage version file
echo -e "${YELLOW}Staging version file...${NC}"
git add "$VERSION_FILE"

# Create release commit
echo -e "${YELLOW}Creating release commit...${NC}"
git commit -m "release: prepare ${VERSION}"

# Create annotated tag
echo -e "${YELLOW}Creating annotated tag...${NC}"
git tag -a "$VERSION" -m "Terraform ContextForge Provider ${VERSION}"

# Verify goreleaser is installed
if ! command -v goreleaser &> /dev/null; then
    echo -e "${RED}Error: goreleaser not found${NC}"
    echo "Install with: go install github.com/goreleaser/goreleaser/v2@latest"
    echo ""
    echo "Rolling back changes..."
    git tag -d "$VERSION"
    git reset --hard HEAD~1
    exit 1
fi

# Check for GITHUB_TOKEN (warn but allow continuation)
if [ -z "$GITHUB_TOKEN" ]; then
    echo -e "${YELLOW}Warning: GITHUB_TOKEN not set${NC}"
    echo "GitHub release creation may fail. Set with:"
    echo "  export GITHUB_TOKEN=your_token_here"
    echo ""
fi

# Check for GPG_FINGERPRINT (required for signing)
if [ -z "$GPG_FINGERPRINT" ]; then
    echo -e "${YELLOW}Warning: GPG_FINGERPRINT not set${NC}"
    echo "Provider signing will fail. Set with:"
    echo "  export GPG_FINGERPRINT=your_gpg_fingerprint"
    echo ""
fi

# Run goreleaser
echo -e "${YELLOW}Running goreleaser...${NC}"
if ! goreleaser release --clean; then
    echo -e "${RED}Error: goreleaser failed${NC}"
    echo ""
    echo "Rolling back changes..."
    git tag -d "$VERSION"
    git reset --hard HEAD~1
    exit 1
fi

# Merge changelog from goreleaser output
if ! merge_changelog "$VERSION"; then
    echo -e "${RED}Error: Failed to merge changelog${NC}"
    echo ""
    echo "Rolling back changes..."
    git tag -d "$VERSION"
    git reset --hard HEAD~1
    exit 1
fi

# Stage and amend commit with changelog
echo -e "${YELLOW}Amending commit with changelog...${NC}"
git add "$CHANGELOG"
git commit --amend --no-edit

# Update tag to point to amended commit
echo -e "${YELLOW}Updating tag...${NC}"
git tag -fa "$VERSION" -m "Terraform ContextForge Provider ${VERSION}"

echo ""
echo -e "${GREEN}Release ${VERSION} prepared successfully!${NC}"
echo ""
echo "Next steps:"
echo "  1. Review changes:"
echo "     git show HEAD"
echo "     git diff HEAD~1 CHANGELOG.md"
echo ""
echo "  2. Review draft release on GitHub:"
echo "     https://github.com/leefowlercu/terraform-provider-contextforge/releases"
echo ""
echo "  3. If everything looks good:"
echo "     git push && git push --tags"
echo ""
echo "  4. Publish draft release on GitHub web UI"
echo ""
echo "  5. If you need to undo:"
echo "     git tag -d ${VERSION} && git reset --hard HEAD~1"
