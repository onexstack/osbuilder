

## Pre-release 版本生成规则

osbuilder semver tag --silent --next  --patch-types refactor --prerelease-mode always
上述命令会在当前prerelease + 1。如果当前没有 prerelease，在patch + 1，然后在alpha.1
