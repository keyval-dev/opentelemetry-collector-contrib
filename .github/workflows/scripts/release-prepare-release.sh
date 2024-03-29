#!/bin/bash -ex

make chlog-update VERSION="v${CANDIDATE_BETA}"
git config user.name opentelemetrybot
git config user.email 107717825+opentelemetrybot@users.noreply.github.com

BRANCH="prepare-release-prs/${CANDIDATE_BETA}"
git checkout -b "${BRANCH}"
git add --all
git commit -m "changelog update ${CANDIDATE_BETA}"

sed -i.bak "s/${CURRENT_BETA}/${CANDIDATE_BETA}/g" versions.yaml
find . -name "*.bak" -type f -delete
git add versions.yaml
git commit -m "update version.yaml ${CANDIDATE_BETA}"

make multimod-prerelease
git add .
git commit -m "make multimod-prerelease changes ${CANDIDATE_BETA}" || (echo "no multimod changes to commit")

make multimod-sync
git add .
git commit -m "make multimod-sync changes ${CANDIDATE_BETA}" || (echo "no multimod changes to commit")

make otelcontribcol

git push origin "${BRANCH}"

gh pr create --title "[chore] Prepare release ${CANDIDATE_BETA}" --body "
The following commands were run to prepare this release:
- make chlog-update VERSION=v${CANDIDATE_BETA}
- sed -i.bak s/${CURRENT_BETA}/${CANDIDATE_BETA}/g versions.yaml
- make multimod-prerelease
- make multimod-sync
"
