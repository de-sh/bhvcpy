from django.apps import AppConfig
from django.conf import settings


class SrvConfig(AppConfig):
    name = "srv"

    def read(self):
        from . import scheduler

        if settings.SCHEDULER_AUTOSTART:
            scheduler.start()
