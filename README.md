# GitHub Search Service (gRPC)

## Overview

This repository contains a gRPC service that acts as a proxy for searching code on GitHub. It leverages the GitHub Search API to perform code queries, providing a gRPC interface for clients to interact with GitHub's search functionality. The service allows searching for code with a given phrase and supports optional filtering by user. It returns the file URL and the repository it was found in for each search result.

## Features

* **gRPC Interface:** Provides a gRPC API for code searches on GitHub.
* **Search Functionality:** Implements the core search functionality using the GitHub Search API.
* **Search Parameters:**
    * `search_term`: The phrase to search for in the code.
    * `user` (optional): Filters the search to a specific user.
* **Result Formatting:** Returns search results with file URLs and repository names.
* **Metadata Handling:** Supports passing optional search parameters via gRPC metadata:
    * `sort`: Sorts the search results ("indexed")
    * `order`: Order of results ("asc" or "desc")
    * `per_page`: Number of results per page (1-100)
    * `page`: Page number of results
* **Input Validation:** Validates the optional search parameters from metadata to ensure they adhere to GitHub API constraints.
* **Error Handling:** Implements gRPC error handling to provide informative error messages to clients.

## API Specification

The service implements the following gRPC API:

```protobuf
service GithubSearchService {
  rpc Search (SearchRequest) returns (SearchResponse);
}

message SearchRequest {
  required string search_term = 1;
  required string user = 2;
}

message SearchResponse {
  repeated Result results = 1;
}

message Result {
  required string file_url = 1;
  required string repo = 2;
}
```
## Search RPC
- **Description:** Performs a code search on GitHub.
- **Request:** SearchRequest message containing the search_term and optional user.
- **Response:** SearchResponse message containing a list of Result messages.

## Implementation Details

* **GitHub API Usage:** The service uses the GitHub Search Code API: `https://docs.github.com/en/rest/search/search?apiVersion=2022-11-28#search-code`
* **Metadata Processing:** gRPC metadata is used to pass optional search parameters to the GitHub API.
* **Validation:** Input validation is performed on the server-side to ensure that optional parameters adhere to the GitHub API's requirements.
* **Error Handling:** gRPC error codes and messages are used to communicate errors to the client.

## Dependencies

* go modules

## Installation

1.  Clone the repository:

    ```bash
    git clone [https://github.com/Pratham700/github-search-service.git](https://github.com/Pratham700/github-search-service.git)
    ```

2.  Navigate to the project directory:

    ```bash
    cd github-search-service
    ```

3.  Build the gRPC server:

    ```bash
    make build-server
    ```

## Usage

1.  Start the gRPC server:

    ```bash
    make run-server
    ```

2.  Clients can then make gRPC calls to the `GithubSearchService` to search for code.

    * Clients need to implement the gRPC client code as per the proto definition.
    * Optional search parameters (`sort`, `order`, `per_page`, `page`) can be sent as metadata with the gRPC request.

## Author

* Pratham700
