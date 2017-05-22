# funnel [![Build Status](https://travis-ci.org/agnivade/funnel.svg?branch=master)](https://travis-ci.org/agnivade/funnel) [![Go Report Card](https://goreportcard.com/badge/github.com/agnivade/funnel)](https://goreportcard.com/report/github.com/agnivade/funnel) [![codecov](https://codecov.io/gh/agnivade/funnel/branch/master/graph/badge.svg)](https://codecov.io/gh/agnivade/funnel) [![Gitter](https://badges.gitter.im/agnivade/funnel.svg)](https://gitter.im/agnivade/funnel)

### A new approach to logging

The 12 factor [rule](https://12factor.net/logs) for logging says that an app "should not attempt to write to or manage logfiles. Instead, each running process writes its event stream, unbuffered, to stdout." The execution environment should take care of capturing the logs and perform further processing with it. Funnel *is* this "execution environment".

All you have to do from your app is to print your log line to stdout, and pipe it to funnel. You can still use any logging library inside your app to handle other stuff like log level, structured logging etc. But don't bother about the log destination. Let funnel take care whether you want to just write to files or stream your output to Kafka. Think of it as a fluentd/logstash replacement(with minimal features!) but having only stdin as an input.

### Features quick tour
- Basic use case of logging to local files:
  * Rolling over to a new file
  * Deleting old files
  * Gzipping files
  * File rename policies
- Prepend each log line with a custom string
- Supports other target outputs like Kafka, ElasticSearch. More info below.
- Live reloading of config on file save. No more messing around with SIGHUP or SIGUSR1.

### Quickstart

Grab the binary for your platform and the config file from [here](https://github.com/agnivade/funnel/releases).

To run, just pipe the output of your app to the funnel binary. Note that, funnel only consumes from stdin, so you might need to redirect stderr to stdout.

```bash
$/etc/myapp/bin 2>&1 | funnel
```

P.S. You also need to drop the funnel binary to your $PATH.

### Target outputs and Use cases

| Output  | Description | Log format  |
|-------- | ----------- | ----------- |
| <img src="http://www.iconsdb.com/icons/preview/black/blank-file-xxl.png" height="32" width="32" style="vertical-align: bottom;" /> File | Writes to local files | No format needed. |
| <img src="https://static.woopra.com/apps/kafka/images/icon-256.png" height="32" width="32" style="vertical-align: bottom;" /> Kafka | Send your log stream to a Kafka topic | No format needed.  |
| <img src="https://cdn4.iconfinder.com/data/icons/redis-2/1451/Untitled-2-32.png" height="32" width="32" style="vertical-align: bottom;" /> Redis pub-sub | Send your log stream to a Redis pub-sub channel | No format needed. |
| <img src="https://nr-platform.s3.amazonaws.com/uploads/platform/published_extension/branding_icon/134/logo.png" height="32" width="32" style="vertical-align: bottom;" /> ElasticSearch | Index, Search and Analyze structured JSON logs | Logs have to be in JSON format |
| <img src="https://nr-platform.s3.amazonaws.com/uploads/platform/published_extension/branding_icon/275/AmazonS3.png" height="32" width="32" /> Amazon S3 | Upload your logs to S3 | No format needed. |
| <img src="https://s-media-cache-ak0.pinimg.com/236x/6c/71/45/6c71456fbd7fca223bb08194a35eeb74.jpg" height="32" width="32" style="vertical-align: bottom;" /> InfluxDB | Use InfluxDB if your app emits timeseries data which needs to be queried and graphed | Logs have to be in JSON format with `tags` and `fields` as the keys |
| <img src="https://nats.io/img/logo.png" height="32" width="32" /> NATS| Send your log stream to a NATS subject | No format needed.

Further details on input log format along with examples can be found in the sample config [file](config.toml#L49).

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

### Disabling outputs

In the case that you don't intend to use the Elasticsearch, InfluxDB, Kafka, Redis or S3 features, e.g. you just want to use the log rotation features, you can reduce the size of the binary by using build tags.

The build tags are:
- `disableelasticsearch`
- `disableinfluxdb`
- `disablekafka`
- `disableredis`
- `disables3`
- `disablenats`

e.g., to build without any of the above outputs:

```
go build -tags "disableelasticsearch disableinfluxdb disablekafka disableredis disables3 disablenats" ./cmd/funnel
```

### Windows Support:

Funnel logs its internal errors to syslog. As syslog is not supported on windows, funnel doesn't work on windows -
https://golang.org/pkg/log/syslog/#pkg-note-BUG

To make it work on windows, just replace the syslog writer with another logging library which atleast has an `Err()` method implemented. 

#### Footnote - This project was heavily inspired from the [logsend](https://github.com/ezotrank/logsend) project.
