# Example: Dockerfile

This directory contains an example of using murmur inside a Dockerfile.

To run the example:

```bash
docker build -t murmur-example .
docker run \
    -e SECRET_SAUCE=passthrough:szechuan \
    murmur-example
```
