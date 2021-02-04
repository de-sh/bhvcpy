from django.apps import AppConfig


class SrvConfig(AppConfig):
    name = 'srv'

    def read(self):
        from . import scheduler
        if settings.SCHEDULER_AUTOSTART:
            scheduler.start()