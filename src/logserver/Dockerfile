FROM tensorflow/tensorflow

RUN apt-get update -y && \
    apt-get install -y git && \
    rm -rf /var/lib/apt/lists/*
ADD . /logserver
RUN python -m pip install --upgrade pip
RUN pip install -r logserver/requirements.txt
ENTRYPOINT [ "python", "-m", "logserver.main" ]
