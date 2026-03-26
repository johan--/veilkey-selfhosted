uuid:           "vk-verifier-20260326"
name:           "verifier"
version:        "1.0.0"
player:         "claude"
player_version: "go"
role:           "member"
skills:         ["skills/go-ci", "skills/rust-ci", "skills/security-audit", "skills/code-review"]
hooks:          []
git: {
	user:         "dal-verifier"
	email:        "dal-verifier@dalcenter.local"
	github_token: "env:GITHUB_TOKEN"
}
