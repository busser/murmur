FROM alpine

COPY --from=ghcr.io/busser/murmur:latest /murmur /bin/murmur

ENTRYPOINT ["/bin/murmur", "run", "--", "sh", "-c", "echo The secret sauce is $SECRET_SAUCE."]
