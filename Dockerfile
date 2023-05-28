FROM scratch

LABEL org.opencontainers.image.source=https://github.com/busser/murmur

# The binary is built beforehand.
COPY murmur /

ENTRYPOINT ["/murmur"]
