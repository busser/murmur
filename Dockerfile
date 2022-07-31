FROM scratch

LABEL org.opencontainers.image.source=https://github.com/busser/whisper

# The binary is built beforehand.
COPY whisper /

ENTRYPOINT ["/whisper"]
