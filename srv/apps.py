from django.apps import AppConfig
from django.conf import settings
from . import scheduler


class SrvConfig(AppConfig):
    name = "srv"

    def read(self):
        if settings.SCHEDULER_AUTOSTART:
            scheduler.start()
