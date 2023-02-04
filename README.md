# ECK Elasticsearch Cluster ERU Limit Exporter

## Description


## Usage

Running the server:
```
./eck-escluster-eru-limit-exporter -conf /path/to/config.yaml [-debug] [-version] [-validate]
```

## Command Flags

`-conf` : The path to the configuration file to be used
`-debug` : Enable debug mode
`-version` : Show version and exit
`-validate` : Validate configuration and exit

*A sample configuration can be found in the `examples/` directory.*

## Available endpoints

* **GET /cluster-limit?cluster=<CLUSTER_NAME>**
    * Return the ERU limit for the specified `cluster` in bytes
* **GET /metrics**
    * Return the list of prometheus metrics for the exporter
* **GET /healthz**
    *  Return the current health status of the exporter
* **GET /config**
    * Return the current config which has been used to start the exporter
* **GET /debug/profile**
    * Generate a debugging profile.  See [here](https://go.dev/blog/pprof) for more details.


## Building

### 1. Checkout required code version

First, ensure you have checked out the proper release tag in order to get all files/dependencies corresponding to that version. 

### 2. Build Go binary

Run `make build` to build the the binary for the current operatory system or run `make build-all` to build for both Linux and OSX.   Refer to the makefile for additional options.

### 3. Build Docker container
Run the following docker command to build the image
```
docker build -t eck-escluster-eru-limit-exporter:$(cat VERSION.txt) --build-arg VERSION=$(cat VERSION.txt) .
```


## License

Covered under the [MIT license](LICENSE.md).

## Author

Alain Lefebvre <hartfordfive 'at' gmail.com>
