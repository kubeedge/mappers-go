FROM python:3.8

ADD ./device /device

WORKDIR /device

ENV PYTHONUNBUFFERED=1

RUN pip install -i  http://pypi.douban.com/simple/ --trusted-host pypi.douban.com -r requirements.txt

CMD ["python", "/device/server.py"]