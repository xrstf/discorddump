# DiscordDump

This program can be used to dump all messages in all guilds that the configured
user has access to. Each channel will result in a single large logfile with one
message (as JSON) per line.

## Usage

    go get github.com/xrstf/discorddump

Create a config file (look at the `config.yaml.dist`) and then run the program
like:

    ./discorddump myconfig.yaml

To only go back to a certain point in time, add the `-cutoff` parameter with a
`YYYY-MM-DD` value:

    ./discorddump -cutoff 2018-10-10 myconfig.yaml
