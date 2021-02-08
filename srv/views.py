from django.shortcuts import render
from django.views.generic import View
from django.http import JsonResponse
from django.conf import settings
import redis


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
            entries = [
                {
                    "code": str(self.redis.lindex("code", i)),
                    "name": str(self.redis.lindex("name", i)),
                    "open": str(self.redis.lindex("open", i)),
                    "high": str(self.redis.lindex("high", i)),
                    "low": str(self.redis.lindex("low", i)),
                    "close": str(self.redis.lindex("close", i)),
                }
                for i in range(int(self.redis.llen("open"))-1, -1, -1)
            ]
            return JsonResponse({"entries": entries}, status=200)

        return render(request, "srv/bhavcopy.html", context={"version": "0.0.1"})
