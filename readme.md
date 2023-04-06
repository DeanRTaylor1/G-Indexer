[![test-ci](https://github.com/DeanRTaylor1/gosearch/actions/workflows/test-ci.yml/badge.svg)](https://github.com/DeanRTaylor1/gosearch/actions/workflows/test-ci.yml)

# GoSearch

GoSearch is a search engine for static websites, implemented in Go. It utilizes the BM25 algorithm to rank search results and provides a simple web interface for user interaction.

## Features

- Web crawler for static websites
- BM25 algorithm for search result ranking
- Web server with a basic user interface for search

## Installation

To install GoSearch, you need to have [Go](https://golang.org/doc/install) installed on your system. Once Go is installed, you can clone this repository:

```bash
git clone https://github.com/DeanRTaylor1/gosearch.git
```

Then, navigate to the project directory and build the project:

```bash
cd gosearch
go build -o ./bin/gosearch .
```

## Usage

After building the project, you can run the gosearch binary with the following command:

```bash
./bin/gosearch
By default, the web server will start on port 8080. Open a web browser and navigate to http://localhost:8080 to use the search interface.
```

You can also use the command-line interface to interact with the search engine. Run ./gosearch --help for more information on available commands and options.

## Contributing

Contributions to GoSearch are welcome! If you have a feature request, bug report, or want to contribute code, please open an issue or create a pull request.

## License

GoSearch is released under the MIT License.
