import logging

from apscheduler.schedulers.background import BackgroundScheduler
from django_apscheduler.jobstores import register_events
from pytz import timezone

from django.conf import settings

from .download import DownloadCSV


def start():
    if settings.DEBUG:
        logging.basicConfig()
        logging.getLogger("apscheduler").setLevel(logging.DEBUG)

    # Create background scheduler on IST
    scheduler = BackgroundScheduler()
    scheduler.configure(timezone=timezone("Asia/Kolkata"))

    downloader = DownloadCSV()

    # Run daily(best case: only on market days), at 6PM
    scheduler.add_job(downloader.daily_bhavcopy, "cron", id="daily_bhavcopy", hour=18)
    register_events(scheduler)
    scheduler.start()