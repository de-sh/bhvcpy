# bhavco.py

### TODO
- [x] Create script to download ZIP, extract CSV and add contents to redis
    - [x] Download ZIP for date
    - [x] Exctract contents of CSV
- [x] Create scheduled scraping policy to update redis at 6PM daily
- [x] Create django webserver to serve Vue-js front-end
    - [x] Basic Django backend serving html
    - [x] Create Vue js front-end with bootstrap UI
    - [x] Use axios to load data
    - [x] Add filtered searching ability to app

### Installation
1. redis-server, Python3.8, Pip are required. If you have pip installed, simply use the following to ensure pipenv is also installed.
```
python3.8 --m pip install pipenv
```
2. install and activate the pipenv virtual environment inside the current directory (`cd bhavco.py`)
```
pipenv install
pipenv shell
```
4. Setup redis-server to run simulataneously in the background. Redis should be accessible at localhost:6379.
```
redis-server &
```
5. Set SECRET_KEY and make migrations of django models.
```
export SECRET_KEY='secret_key_here'
./manage.py makemigrations
./manage.py migrate
```
6. Run server on port 8000
```
./manage.py runserver --noreload 0.0.0.0:8000
```
7. Visit http://localhost:8000 and interact with website. This might be empty if redis hasn't been updated by the background Bhavcopy downloader task setup with APScheduler.
8. Close website and redis instances.

### Redis schema
- "codes": List of unique code for each and every scrip traded on the market.
- "open", "high", "low", "close": Also insert ordered listing scrip details.

## Possible Optimization
I believe the application presented in this repository is extremely crude and could require optimizations such as streaming of csv data as small packets.
