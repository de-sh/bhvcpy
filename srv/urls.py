from django.urls import path
from .views import SrView

urlpatterns = [
    path('', SrView.as_view(), name='srv_url'),
]
