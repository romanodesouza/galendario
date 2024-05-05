#!/bin/bash

set -e

# Install
go install ./cmd/galendario

# Fetch new calendar
galendario \
	| sed 's,\r,,g' \
	| while read r; do echo -n "$r\r\n"; done > galendario_inline_new.ics

# Fetch current calendar
curl \
	-s \
	-H "Accept: application/vnd.github+json" \
	-H "Authorization: Bearer ${GH_BEARER}" \
	-H "X-GitHub-Api-Version: 2022-11-28" \
	https://api.github.com/gists/${GH_GIST} \
	| jq '.files["galendario.ics"]["content"]' \
	| sed 's,",,g' \
	| tr -d '\n' > galendario_inline_current.ics

# Diff
diff galendario_inline_new.ics galendario_inline_current.ics > /dev/null && echo "No updates" && exit 0

# Update
NEW_CONTENT=$(cat galendario_inline_new.ics)
curl \
	-s \
	-X PATCH \
	-H "Accept: application/vnd.github+json" \
	-H "Authorization: Bearer ${GH_BEARER}" \
	-H "X-GitHub-Api-Version: 2022-11-28" \
	https://api.github.com/gists/${GH_GIST} \
	-d "{\"description\":\"Update\",\"files\":{\"galendario.ics\":{\"content\":\"${NEW_CONTENT}\"}}}" > /dev/null

echo "Calendar has been updated"
