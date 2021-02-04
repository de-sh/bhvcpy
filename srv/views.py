from django.shortcuts import render
from django.views.generic import View
from django.http import JsonResponse
from django.conf import settings
import redis

""" from .models import Srv """


class SrView(View):
    def __init__(self):
        self.redis = redis.StrictRedis(
            host=settings.REDIS_HOST, port=settings.REDIS_PORT, db=0
        )

    def get(self, request):
        entries = [
            {
                "code": self.redis.lindex("code", i),
                "name": self.redis.lindex("name", i),
                "open": self.redis.lindex("open", i),
                "high": self.redis.lindex("high", i),
                "low": self.redis.lindex("low", i),
                "close": self.redis.lindex("close", i),
            }
            for i in range(0, self.redis.llen("code"))
        ]

        if request.is_ajax():
            return JsonResponse({"entries": entries}, status=200)

        return render(request, "srv/bhavcopy.html", context={"version": "0.0.1"})
