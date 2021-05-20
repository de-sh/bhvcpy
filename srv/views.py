from django.shortcuts import render
from django.views.generic import View
from django.http import JsonResponse
from django.conf import settings
import redis
import json


class SrView(View):
    def __init__(self):
        self.redis = redis.StrictRedis(
            host=settings.REDIS_HOST,
            port=settings.REDIS_PORT,
            decode_responses=True,
            db=0,
        )

    def get(self, request):
        if request.is_ajax():
            names = [str(self.redis.lindex("name", i))
                     for i in range(int(self.redis.llen("name"))-1, -1, -1)]
            if "key" in request.GET:
                key = request.GET["key"].upper()
                names = list(filter(lambda x: x.find(key) != -1, names))
            entries = [self.redis.hgetall(name) for name in names]
            return JsonResponse({"entries": entries}, status=200)

        return render(request, "srv/bhavcopy.html", context={"version": "0.0.1"})
