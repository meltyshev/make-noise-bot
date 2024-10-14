import os

from bmemcached import Client

mc = Client(
    (os.environ['MEMCACHED_SERVER'],)
)
