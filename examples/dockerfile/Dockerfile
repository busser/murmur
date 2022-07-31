FROM alpine

COPY --from=ghcr.io/busser/whisper:latest /whisper /bin/whisper

ENTRYPOINT ["/bin/whisper", "exec", "--", "sh", "-c", "echo The secret sauce is $SECRET_SAUCE."]
