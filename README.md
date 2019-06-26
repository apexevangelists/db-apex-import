# db-apex-import

Command to connect to an Oracle Database and import an Application Express application

## Prerequisites

Oracle Instant Client must already be installed

[Oracle Instant Client](https://www.oracle.com/database/technologies/instant-client.html)

Note - Oracle Instant Client must be configured per your environment (please follow the instructions provided by Oracle).

## Table of Contents

- [db-apex-import](#db-apex-import)
  - [Prerequisites](#Prerequisites)
  - [Table of Contents](#Table-of-Contents)
  - [Installation](#Installation)
  - [Building](#Building)
  - [Usage](#Usage)
  - [Support](#Support)
  - [Contributing](#Contributing)

## Installation

1) Clone this repository into a local directory, copy the db-apex-import executable into your $PATH

```bash
$ git clone https://github.com/apexevangelists/db-apex-import
```

## Building

Pre-requisite - install Go

Compile the program -

```bash
$ go build
```

## Usage

```bash-3.2$ ./db-apex-import -h
Usage of ./db-apex-import:
  -alias string
    	application alias (override)
  -appID string
    	application ID to import into
  -configFile string
    	Configuration file for general parameters (default "config")
  -connection string
    	Configuration file for connection
  -db string
    	Database Connection, e.g. user/password@host:port/sid
  -debug
    	Debug mode (default=false)
  -i string
    	Script File to Import
  -schema string
    	schema to import into (override)
  -workspace string
    	workspace to import into (override)

bash-3.2$
```

## Support

Please [open an issue](https://github.com/apexevangelists/db-apex-import/issues/new) for support.

## Contributing

Please contribute using [Github Flow](https://guides.github.com/introduction/flow/). Create a branch, add commits, and [open a pull request](https://github.com/apexevangelists/db-apex-import/compare).