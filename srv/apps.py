from django.apps import AppConfig
from django.conf import settings


class SrvConfig(AppConfig):
    name = "srv"
    run_already = False

    def ready(self):
        if SrvConfig.run_already:
            return

        from . import scheduler

        if settings.SCHEDULER_AUTOSTART:
            scheduler.start()
            SrvConfig.run_already = True
