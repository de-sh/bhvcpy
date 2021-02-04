from django.shortcuts import render
from django.views.generic import View
from django.http import JsonResponse

from .models import Srv

class SrView(View):
    def get(self, request):
        entries = list(Srv.objects.values())

        if request.is_ajax():
            return JsonResponse({'entries': entries}, status = 200)

        return render(request, 'srv/bhavcopy.html', context={'version': '0.0.1'})