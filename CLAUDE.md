# Bash permissions
When chaining multiple commands with &&, ;, or |, always wrap them in `bash -c '...'` so they match the allowed `Bash(bash *)` permission pattern.
