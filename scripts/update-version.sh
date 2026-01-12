#/bin/bash

function file_exists() {
    FILENAME=$1
    if [[ ! -f "$FILENAME" ]]; then
        echo "Error: $FILENAME not found. Cannot continue"
        exit 1
    fi
}

function backup_file() {
    local SRC=$1
    echo "Backing up $SRC as $SRC.bak"
    cp "$SRC" "$SRC.bak"
    if [ $? -ne 0 ]; then
        echo "Failed to backup $SRC. Cannot continue"
        exit 1
    fi
}

function execute_sed() {
    local FILENAME=$1
    local SEARCH=$2
    local REPLACE=$3
    local VERIFICATION=$4
    /usr/bin/sed -i "" "s/$SEARCH/$REPLACE/" $FILENAME
    if [ $? -ne 0 ]; then
        echo "Failed to update $A1. Cannot continue"
        exit 1
    fi

    echo "Verifying the changes to $FILENAME"
    /usr/bin/grep -E "$VERIFICATION" $FILENAME 1>/dev/null
    if [ $? -ne 0 ]; then
        echo "Failed to update $FILENAME. Cannot continue"
        exit 1
    fi
}

function modify_makefile() {
    local FILENAME=$1
    local NEW_VERSION=$2
    local SEARCH="^PSTREE_VERSION := [0-9\.]*"
    local REPLACE="PSTREE_VERSION := $NEW_VERSION"
    local VERIFICATION="^PSTREE_VERSION\s*:=\s*$NEW_VERSION"
    echo "Updating PSTREE_VERSION in $FILENAME to $NEW_VERSION"
    execute_sed "$FILENAME" "$SEARCH" "$REPLACE" "$VERIFICATION"
}

function modify_root_cmd() {
    local FILENAME=$1
    local NEW_VERSION=$2
    local SEARCH="^[[:space:]]*version[[:space:]]*string[[:space:]]*=[[:space:]]*\"[0-9\.]*\""
    local REPLACE="\tversion               string = \"$NEW_VERSION\""
    local VERIFICATION="^[[:space:]]*version[[:space:]]*string[[:space:]]*=[[:space:]]*\"$NEW_VERSION\""
    echo "Updating version in $FILENAME to $NEW_VERSION"
    execute_sed "$FILENAME" "$SEARCH" "$REPLACE" "$VERIFICATION"
    echo Reformatting $FILENAME
    go fmt $FILENAME
}

function modify_manpage() {
    local FILENAME=$1
    local NEW_VERSION=$2
    local NEWDATE=$(date +"%B %d, %Y")
    local SEARCH="^.TH PSTREE 1 \".*\" \".*\" \"User Commands\""
    local REPLACE=".TH PSTREE 1 \"$NEWDATE\" \"$NEW_VERSION\" \"User Commands\""
    local VERIFICATION="^.TH PSTREE 1 \"$NEWDATE\" \"$NEW_VERSION\""
    echo "Updating data and version in $FILENAME"
    execute_sed "$FILENAME" "$SEARCH" "$REPLACE" "$VERIFICATION"
    echo "Generating the HTML version of the man page"
    groff -Thtml -mandoc "$FILENAME" > "${REPO_ROOT}/doc/pstree.1.html"
}

function modify_readme() {
    local FILENAME=$1
    local NEW_VERSION=$2
    local SEARCH="pstree .Revision: [0-9\.]* . by Cursed Bananazon (C) 2025, 2026"
    local REPLACE="pstree \$Revision: $NEW_VERSION \$ by Cursed Bananazon \(C\) 2025, 2026"
    local VERIFICATION="pstree .Revision: $NEW_VERSION . by Cursed Bananazon \(C\) 2025, 2026"
    echo "Updating PSTREE_VERSION in $FILENAME to $NEW_VERSION"
    execute_sed "$FILENAME" "$SEARCH" "$REPLACE" "$VERIFICATION"
}

function modify_html() {
    # groff -Thtml -mandoc foo.1 > foo.html
    local MANPAGE=$1
    local FILENAME=$2
    local NEW_VERSION=$3
    local SEARCH="^<title>PSTREE 1"
    local REPLACE="<title>PSTREE 1 $NEW_VERSION"
    local VERIFICATION="^<title>PSTREE 1 $NEW_VERSION"
    echo "Updating $FILENAME"
    groff -Thtml -mandoc $MANPAGE > $FILENAME
}

REPO_ROOT=$(git rev-parse --show-toplevel)
if [ -z $REPO_ROOT ]; then
    echo "Failed to determine the repo root. Cannot continue."
    exit 1
fi

if [ -z $PSTREE_VERSION ]; then
    echo "PSTREE_VERSION is not set. Cannot continue."
    exit 1
fi

# ${REPO_ROOT}/Makefile
MAKEFILE="${REPO_ROOT}/Makefile"
file_exists $MAKEFILE
backup_file $MAKEFILE
modify_makefile $MAKEFILE $PSTREE_VERSION
echo

# ${REPO_ROOT}/cmd/root.go
ROOT_CMD="${REPO_ROOT}/cmd/root.go"
file_exists $ROOT_CMD
backup_file $ROOT_CMD
modify_root_cmd $ROOT_CMD $PSTREE_VERSION
echo

# ${REPO_ROOT}/share/man/man1/pstree.1
MANPAGE="${REPO_ROOT}/share/man/man1/pstree.1"
file_exists $MANPAGE
backup_file $MANPAGE
modify_manpage $MANPAGE $PSTREE_VERSION
echo


# ${REPO_ROOT}/README.md
README="${REPO_ROOT}/README.md"
file_exists $README
backup_file $README
modify_readme $README $PSTREE_VERSION
echo

# ${REPO_ROOT}/doc/pstree.1.html
HTML="${REPO_ROOT}/doc/pstree.1.html"
file_exists $HTML
backup_file $HTML
modify_html $MANPAGE $HTML $PSTREE_VERSION
echo
