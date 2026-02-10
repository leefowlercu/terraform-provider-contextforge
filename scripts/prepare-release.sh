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
REPO_COMPARE_BASE="https://github.com/leefowlercu/terraform-provider-contextforge/compare"
REPO_RELEASE_BASE="https://github.com/leefowlercu/terraform-provider-contextforge/releases/tag"

# Cleanup function to remove temp files on exit
cleanup() {
    rm -f "${PROJECT_ROOT}/.next-version"
}
trap cleanup EXIT

# Upsert a markdown reference link in a file.
upsert_markdown_link() {
    local KEY=$1
    local VALUE=$2
    local FILE=$3
    local TMP_FILE

    TMP_FILE=$(mktemp)
    awk -v key="$KEY" -v value="$VALUE" '
        BEGIN { replaced = 0 }
        index($0, "[" key "]:") == 1 {
            print value
            replaced = 1
            next
        }
        { print }
        END {
            if (replaced == 0) {
                print value
            }
        }
    ' "$FILE" > "$TMP_FILE"
    mv "$TMP_FILE" "$FILE"
}

# Roll back local tag/commit changes created by this script before publish.
rollback_pre_publish_changes() {
    local VERSION=$1
    local PREPARED_COMMIT=$2

    echo ""
    echo "Rolling back local release preparation changes..."
    git tag -d "$VERSION" 2>/dev/null || true

    if [ "$PREPARED_COMMIT" = true ]; then
        # Keep local file safety checks; this script is expected to run from a clean tree.
        git reset --keep HEAD~1 2>/dev/null || true
    fi
}

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
                echo "" >> "$TEMP_CHANGELOG"
                echo "### $SECTION_NAME" >> "$TEMP_CHANGELOG"
                echo "" >> "$TEMP_CHANGELOG"
                echo -e "$ENTRIES" >> "$TEMP_CHANGELOG"
            fi

            SECTION_NAME="${BASH_REMATCH[1]}"
            ENTRIES=""
            IN_SECTION=true
        # Check for commit entries (e.g., "* abc1234 commit message")
        elif [[ "$line" =~ ^\*[[:space:]]+[a-f0-9]{7}[[:space:]]+(.*) ]]; then
            local MESSAGE="${BASH_REMATCH[1]}"
            # Capitalize first letter
            local FIRST_CHAR=$(echo "${MESSAGE:0:1}" | tr '[:lower:]' '[:upper:]')
            local REST="${MESSAGE:1}"
            MESSAGE="${FIRST_CHAR}${REST}"
            # Add to entries without trailing newline (will be added by echo -e)
            if [ -z "$ENTRIES" ]; then
                ENTRIES="- ${MESSAGE}"
            else
                ENTRIES="${ENTRIES}\n- ${MESSAGE}"
            fi
        fi
    done < "$DIST_CHANGELOG"

    # Save last section
    if [ -n "$ENTRIES" ] && [ -n "$SECTION_NAME" ]; then
        echo "" >> "$TEMP_CHANGELOG"
        echo "### $SECTION_NAME" >> "$TEMP_CHANGELOG"
        echo "" >> "$TEMP_CHANGELOG"
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

    # Build new version section with trailing blank line
    local VERSION_SECTION="## [${VERSION}] - ${TAG_DATE}\n"
    VERSION_SECTION="${VERSION_SECTION}$(cat "$TEMP_CHANGELOG")"
    VERSION_SECTION="${VERSION_SECTION}\n"

    # Find Unreleased section and the next version section
    local UNRELEASED_LINE=$(grep -n "^## \[Unreleased\]" "$CHANGELOG" | cut -d: -f1)
    local NEXT_VERSION_LINE=$(grep -n "^## \[[v0-9]" "$CHANGELOG" | head -1 | cut -d: -f1)

    if [ -n "$UNRELEASED_LINE" ]; then
        # We have an Unreleased section - insert new version after it and clear its content
        {
            # Everything before and including Unreleased header
            head -n "$UNRELEASED_LINE" "$CHANGELOG"
            # Empty line after Unreleased
            echo ""
            # New version section
            echo -e "$VERSION_SECTION"
            # Everything after the Unreleased section content (skip to next version or links)
            if [ -n "$NEXT_VERSION_LINE" ]; then
                tail -n +"$NEXT_VERSION_LINE" "$CHANGELOG"
            else
                # No existing versions, just get the links section
                grep -n "^\[" "$CHANGELOG" | head -1 | cut -d: -f1 | {
                    read LINKS_LINE
                    if [ -n "$LINKS_LINE" ]; then
                        tail -n +"$LINKS_LINE" "$CHANGELOG"
                    fi
                }
            fi
        } > "${CHANGELOG}.tmp"
    else
        # No Unreleased section, insert before first version
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
    fi

    mv "${CHANGELOG}.tmp" "$CHANGELOG"

    # Update version/unreleased reference links.
    local VERSION_LINK=""
    local UNRELEASED_LINK="[Unreleased]: ${REPO_COMPARE_BASE}/${VERSION}...HEAD"

    if [ -n "$PREVIOUS_VERSION" ]; then
        VERSION_LINK="[${VERSION}]: ${REPO_COMPARE_BASE}/${PREVIOUS_VERSION}...${VERSION}"
    else
        VERSION_LINK="[${VERSION}]: ${REPO_RELEASE_BASE}/${VERSION}"
    fi

    upsert_markdown_link "$VERSION" "$VERSION_LINK" "$CHANGELOG"
    upsert_markdown_link "Unreleased" "$UNRELEASED_LINK" "$CHANGELOG"

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
PREVIOUS_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
PREPARED_COMMIT=false
DRY_RUN="${RELEASE_DRY_RUN:-0}"

# Validate version format (vX.Y.Z)
if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    echo -e "${RED}Error: Invalid version format '${VERSION}'${NC}"
    echo "Expected format: vX.Y.Z (e.g., v0.1.0)"
    exit 1
fi

# Remove 'v' prefix for version file
VERSION_NO_V="${VERSION#v}"

echo -e "${GREEN}Preparing release ${VERSION}...${NC}"

# Safety check: release preparation assumes a clean worktree.
if ! git diff-index --quiet HEAD -- || ! git diff-index --cached --quiet HEAD --; then
    echo -e "${RED}Error: Working tree must be clean before running prepare-release.sh${NC}"
    exit 1
fi

# Verify generated provider docs are current before tagging a release.
echo -e "${YELLOW}Verifying Terraform provider docs are up to date...${NC}"
if ! make -C "${PROJECT_ROOT}" docs-check; then
    echo -e "${RED}Error: Provider docs are out of date${NC}"
    echo "Run 'make docs', commit the docs changes, and rerun release preparation."
    exit 1
fi

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

# Track whether we created a commit (for rollback purposes)
CREATED_COMMIT=false

# Check if there are any staged changes
if ! git diff --cached --quiet; then
    # Create release commit only if there are changes
    echo -e "${YELLOW}Creating release commit...${NC}"
    git commit -m "release: prepare ${VERSION}"
    CREATED_COMMIT=true
    PREPARED_COMMIT=true
else
    echo -e "${YELLOW}Version already correct in ${VERSION_FILE}, skipping commit...${NC}"
fi

# Create annotated tag
echo -e "${YELLOW}Creating annotated tag...${NC}"
git tag -a "$VERSION" -m "Terraform ContextForge Provider ${VERSION}"

# Verify goreleaser is installed
if ! command -v goreleaser &> /dev/null; then
    echo -e "${RED}Error: goreleaser not found${NC}"
    echo "Install with: go install github.com/goreleaser/goreleaser/v2@latest"
    rollback_pre_publish_changes "$VERSION" "$PREPARED_COMMIT"
    exit 1
fi

# Generate changelog without publishing so CHANGELOG.md can be committed first.
echo -e "${YELLOW}Generating changelog artifacts (no publish)...${NC}"
if ! goreleaser release --clean --skip=publish,announce; then
    echo -e "${RED}Error: goreleaser changelog generation failed${NC}"
    rollback_pre_publish_changes "$VERSION" "$PREPARED_COMMIT"
    exit 1
fi

# Merge changelog from goreleaser output
if ! merge_changelog "$VERSION"; then
    echo -e "${RED}Error: Failed to merge changelog${NC}"
    rollback_pre_publish_changes "$VERSION" "$PREPARED_COMMIT"
    exit 1
fi

# Stage changelog and create/amend commit
git add "$CHANGELOG"

# Check if changelog was actually modified
if ! git diff --cached --quiet; then
    if [ "$CREATED_COMMIT" = true ]; then
        # Amend existing commit with changelog
        echo -e "${YELLOW}Amending commit with changelog...${NC}"
        git commit --amend --no-edit

        # Update tag to point to amended commit
        echo -e "${YELLOW}Updating tag...${NC}"
        git tag -fa "$VERSION" -m "Terraform ContextForge Provider ${VERSION}"
    else
        # Create new commit with changelog
        echo -e "${YELLOW}Creating release commit with changelog...${NC}"
        git commit -m "release: prepare ${VERSION}"
        PREPARED_COMMIT=true

        # Update tag to point to new commit
        echo -e "${YELLOW}Updating tag...${NC}"
        git tag -fa "$VERSION" -m "Terraform ContextForge Provider ${VERSION}"
    fi
else
    echo -e "${YELLOW}No changelog changes to commit${NC}"
    # Tag is already pointing to correct commit
fi

# Require GitHub token for publish step.
if [ "$DRY_RUN" = "1" ]; then
    echo -e "${YELLOW}Dry run mode enabled (RELEASE_DRY_RUN=1). Skipping goreleaser publish step.${NC}"
else
    if [ -z "$GITHUB_TOKEN" ]; then
        echo -e "${RED}Error: GITHUB_TOKEN not set${NC}"
        echo "Set with: export GITHUB_TOKEN=your_token_here"
        echo "Release preparation is complete locally, but publish was skipped."
        exit 1
    fi

    # Publish release draft only after final commit/tag are in place.
    echo -e "${YELLOW}Publishing draft release with goreleaser...${NC}"
    if ! goreleaser release --clean; then
        echo -e "${RED}Error: goreleaser publish failed${NC}"
        echo "Local release commit/tag remain prepared. Fix the issue and rerun:"
        echo "  goreleaser release --clean"
        exit 1
    fi
fi

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
echo "     git tag -d ${VERSION} && git reset --keep HEAD~1"
