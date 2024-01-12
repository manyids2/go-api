#!/bin/bash

d="${1-.}"
f="${2-NS}"
command='./go-api --path {} --format '"$f"
find "$d" -type f -name "*.go" | fzf \
	--reverse \
	--preview "$command" \
	--preview-window=right,70%
