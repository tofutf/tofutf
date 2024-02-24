FROM alpine:3.19.1

# bubblewrap is for sandboxing, and git permits pulling modules via
# the git protocol
RUN apk add --no-cache bubblewrap git

COPY tofutfd /usr/local/bin/tofutfd

ENTRYPOINT ["/usr/local/bin/tofutfd"]
