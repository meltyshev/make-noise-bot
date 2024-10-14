#!/bin/bash
set -e
exec gunicorn -b 0.0.0.0:8000 app:app & python clock.py $@
