from django.apps import AppConfig
from django.conf import settings


class SrvConfig(AppConfig):
    name = "srv"

    def ready(self):
        from . import scheduler
        if settings.SCHEDULER_AUTOSTART:
            scheduler.start()
