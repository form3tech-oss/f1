#!/bin/sh
set -x

if [ "" != "${TRAVIS_PULL_REQUEST_BRANCH}"  ]; then
    git checkout ${TRAVIS_PULL_REQUEST_BRANCH}
else
    git checkout ${TRAVIS_BRANCH}
fi

# Allow go mod access to proviate github repos
export GOPRIVATE=github.com/form3tech
echo machine github.com login github-build-user password ${GITHUB_TOKEN} > ${HOME}/.netrc

# update vendor directories and push any changes
go mod tidy
go mod vendor

git add go.mod go.sum vendor
git commit --message "vendor update" --no-verify

git remote add origin-auth https://${GITHUB_TOKEN}@github.com/${TRAVIS_REPO_SLUG}.git > /dev/null 2>&1
git push --quiet --set-upstream origin-auth ${TRAVIS_PULL_REQUEST_BRANCH}
