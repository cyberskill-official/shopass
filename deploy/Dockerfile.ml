# Forecast job: reads price history, writes price_forecast. Context = repo root.
FROM python:3.11-slim
RUN apt-get update && apt-get install -y --no-install-recommends build-essential \
    && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY services/ml/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt \
    && python -c "import cmdstanpy; cmdstanpy.install_cmdstan()"
COPY services/ml/ .
ENV PYTHONPATH=/app
ENTRYPOINT ["python", "-m", "bottom.run_forecast"]
