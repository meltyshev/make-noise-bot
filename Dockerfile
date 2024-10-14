FROM python:3.6

WORKDIR /app

COPY requirements.txt .

RUN pip install "setuptools<58.0.0" --no-cache-dir
RUN pip install -r requirements.txt --no-cache-dir

COPY . .
RUN chmod +x start.sh

EXPOSE 8000

CMD ["./start.sh"]
