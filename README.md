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
1. Install python, pip, pipenv as pre-requisites
2. Install django and other project specific requirements with pipenv from within project directory and enter shell.
```
pipenv shell
```
3. Make django migrations done.
```
python manage.py makemigrations
python manage.py migrate
```
4. Run redis and django server simulataneously on local.
```
redis-server &
python manage.py runserver
```
5. Visit url provided by django cli and interact with website.
6. Close django cli and pipenv shell.

### Redis schema
- "codes": List of unique code for each and every scrip traded on the market.
- "open", "high", "low", "close": Also insert ordered listing scrip details.

## Possible Optimization
I believe the application presented in this repository is extremely crude and could require optimizations such as streaming of csv data as small packets.