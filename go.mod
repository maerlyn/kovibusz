module github.com/maerlyn/kovibusz

go 1.13

replace (
	github.com/maerlyn/kovibusz/bkk => ./bkk
	github.com/maerlyn/kovibusz/waze => ./waze
)

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/nlopes/slack v0.6.0
)
