# bhavco.py

### TODO
- [x] Create golang webserver to serve Vue-js front-end
    - [x] Basic `go/http` backend serving html
    - [x] Create Vue js front-end with bootstrap UI
    - [x] Use axios to load data from `/get`
    - [x] Add filtered searching ability to app

### Installation
1. redis-server, >golang-go1.14.0 are required.
2. Setup redis-server to run simulataneously in the background. Redis should be accessible at localhost:6379.
```
redis-server
```
3. Set tempory folder path and start golang server
```
BHVCPY_TEMP="./tmp" go run server.go
```
4. Visit http://localhost:3000 and interact with website. Input a keyword to get values as initially it is empty, this is similarly empty if redis hasn't been updated by the Bhavcopy downloader task.
5. Close website and redis instances.

### Redis schema
- "codes": List of unique code for each and every scrip traded on the market.
- "open", "high", "low", "close": Also insert ordered listing scrip details.
