from django.db import models

# Create your models here.
class Song(models.Model):
    name = models.CharField(max_length=255)
    artist = models.CharField(max_length=255)
    duration = models.IntegerField()
    thumbnail = models.ImageField(upload_to="images/")
    file=models.FileField(upload_to="music/")