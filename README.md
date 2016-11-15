# funnel [![Build Status](https://travis-ci.org/agnivade/funnel.svg?branch=master)](https://travis-ci.org/agnivade/funnel) [![Go Report Card](https://goreportcard.com/badge/github.com/agnivade/funnel)](https://goreportcard.com/report/github.com/agnivade/funnel) [![codecov](https://codecov.io/gh/agnivade/funnel/branch/master/graph/badge.svg)](https://codecov.io/gh/agnivade/funnel)


The 12 factor [rule](https://12factor.net/logs) for logging says that an app "should not attempt to write to or manage logfiles. Instead, each running process writes its event stream, unbuffered, to stdout." The execution environment should take care of capturing the logs and perform further processing with it.

Funnel is meant to be a replacement for your app's "logger + [logrotate](http://www.linuxcommand.org/man_pages/logrotate8.html)" pipeline. No more sending SIGHUP signals, or reload commands to open a new file. No more setting up conditional flags to switch between writing to console and file when changing from dev to production. All you need to do is just print to stdout and pipe it to funnel. And let it take care of the rest.

### Features quick tour
- Basic feature set of a log rotator:
 * Rolling over to a new file
 * Deleting old files
 * Gzipping files
 * File rename policies
- Prepend each log line with a custom string
- Live reloading of config file on save. No more messing around with SIGHUP or SIGUSR1.

### Quickstart

Grab the binary for your platform and the config file from [here](https://github.com/agnivade/funnel/releases).

To run, just pipe the output of your app to the funnel binary. Note that, funnel only consumes from stdin, so you might need to redirect stderr to stdout.

```bash
$/etc/myapp/bin 2>&1 | funnel
```

P.S. You also need to drop the funnel binary to your $PATH.

### Configuration

The config can be specified in a .toml file. The file is part of the repo, which you can see [here](config.toml). All the settings are documented and are populated with the default values. The same defaults are embedded in the app itself, so the app can even run without a config file.

To read the config, the app looks for a file named `config.toml` in these locations one by one -
- `/etc/funnel/config.toml`
- `$HOME/.funnel/config.toml`
- `./config.toml` (i.e. in the current directory of your target app)

You can place a global file in `/etc/funnel/` and have separate files in each app directory to have config values overriding the global ones.

Environment variables are also supported and takes the highest precedence. To get the env variable name, just capitalize the config variable. For eg-
- `logging.directory` becomes `LOGGING_DIRECTORY`
- `rollup.file_rename_policy` becomes `ROLLUP_FILE_RENAME_POLICY`

### TODO:
- Add benchmarks
- Add new output targets like ElasticSearch, InfluxDB, AmazonS3
- Add stats endpoint to expose metrics.

#### Footnote - This project was heavily inspired from the [logsend](https://github.com/ezotrank/logsend) project.
