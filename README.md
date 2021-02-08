# bhavco.py

### TODO
- [ ] Create script to download ZIP, extract CSV and add contents to redis
    - [x] Download ZIP for date
    - [x] Exctract contents of CSV
- [ ] Create scheduled scraping policy to update redis at 6PM daily
- [ ] Create django webserver to serve Vue-js front-end
    - [x] Basic Django backend serving html
    - [x] Create Vue js front-end with bootstrap UI
    - [x] Use axios to load data
    - [ ] Add filtered searching ability to app

### Installation
1. Install docker, redis-server and other pre-requisites
2. Build docker image
```
docker build -t bhvcpy .
```
4. Run redis and web server(docker image) simulataneously. Redis should be accessible at localhost:6379.
```
redis-server &
docker run --net=host -d -p 8000:8000 bhvcpy:latest
```
5. Visit url provided by django cli and interact with website. This might be empty if redis hasn't been updated by Bhavcopy downloader.
6. Close docker and redis instances.

### Redis schema
- "codes": List of unique code for each and every scrip traded on the market.
- "open", "high", "low", "close": Also insert ordered listing scrip details.

## Possible Optimization
I believe the application presented in this repository is extremely crude and could require optimizations such as streaming of csv data as small packets.