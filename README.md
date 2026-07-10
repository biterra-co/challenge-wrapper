# Biterra challenge wrapper

Open-source browser presence telemetry and a Docker parent image for HTTP challenges. Presence is approximate operational analytics, not a billing or security measurement.

```dockerfile
FROM ghcr.io/biterra-co/challenge-wrapper-python:1.0.0
WORKDIR /app
COPY . .
ENV BITERRA_UPSTREAM_PORT=3000
CMD ["python", "app.py"]
```

The inherited entrypoint launches the child `CMD`, listens on port 8080, proxies to the private upstream port, and injects the same-origin browser client into HTML. WebSockets and non-HTML traffic pass through unchanged.
