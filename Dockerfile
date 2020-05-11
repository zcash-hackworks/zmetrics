FROM jupyter/scipy-notebook

ADD --chown=1000 ./data/example-with-bokeh.ipynb  /home/jovyan/
ADD ./data/zcashmetrics.json /home/jovyan/
