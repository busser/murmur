FROM scratch

ARG BINARY=murmur

LABEL org.opencontainers.image.source=https://github.com/busser/murmur

# The binary is built beforehand.
COPY ${BINARY} /

ENTRYPOINT ["/${BINARY}"]
