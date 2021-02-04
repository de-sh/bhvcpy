import requests
import zipfile
import csv 

def download_bhavcopy(date):
    hdr = {
        'User-Agent': 'Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.11 (KHTML, like Gecko) Chrome/23.0.1271.64 Safari/537.11',
        'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
        'Accept-Charset': 'ISO-8859-1,utf-8;q=0.7,*;q=0.3',
        'Accept-Encoding': 'none',
        'Accept-Language': 'en-US,en;q=0.8',
        'Connection': 'keep-alive'
    }
    temp = 'tmp.zip'
    url = f'https://www.bseindia.com/download/BhavCopy/Equity/EQ{date}_CSV.ZIP'

    with requests.get(url, headers=hdr) as r:
        with open(temp, 'wb') as fd:
            for chunk in r.iter_content(chunk_size=128):
                fd.write(chunk)

    with zipfile.ZipFile(temp, 'r') as zf:
        zf.extractall()
        zf.close()

    with open(f'EQ{date}.CSV', 'r') as csv_file:
        csv_reader = csv.reader(csv_file)
        first = True

        for row in csv_reader:
            if first:
                first = False
            else:
                code, name, opn, high, low, close = row[0], row[1], row[4], row[5], row[6], row[7]
                print(code, name, opn, high, low, close)


download_bhavcopy('030221')