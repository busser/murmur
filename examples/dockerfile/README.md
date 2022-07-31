# Example: Dockerfile

This directory contains an example of using whisper inside a Dockerfile.

To run the example:

```bash
docker build -t whisper-example .
docker run \
    -e SECRET_SAUCE=passthrough:szechuan \
    whisper-example
```
