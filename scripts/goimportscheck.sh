#!/usr/bin/env bash

# Check goimports
echo "==> Checking that code complies with goimports requirements..."
goimports_files=$(goimports -l `find . -name '*.go' | grep -v vendor`)
if [[ -n ${goimports_files} ]]; then
    echo 'goimports needs running on the following files:'
    echo "${goimports_files}"
    echo "You can use the command: \`make goimports\` to reformat code."
    exit 1
fi

exit 0