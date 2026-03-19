FROM rust:1-alpine AS builder
RUN apk add --no-cache musl-dev openssl-dev openssl-libs-static perl
WORKDIR /build

# veil-cli
COPY veil-cli/ ./veil-cli/
RUN cd veil-cli && cargo build --release

# veilkey-cli
COPY veilkey-cli/ ./veilkey-cli/
RUN cd veilkey-cli && cargo build --release

FROM alpine:3.20
RUN apk add --no-cache bash curl
COPY --from=builder /build/veil-cli/target/release/veil /usr/local/bin/veil
COPY --from=builder /build/veil-cli/target/release/veilkey /usr/local/bin/veilkey
COPY --from=builder /build/veilkey-cli/target/release/veilkey-cli /usr/local/bin/veilkey-cli
COPY --from=builder /build/veilkey-cli/target/release/veilkey-session-config /usr/local/bin/veilkey-session-config

COPY veilkey-cli/examples/.veilkey.yml /root/.veilkey.yml

ENV VEILKEY_BIN=/usr/local/bin/veilkey
ENV VEILKEY_CLI_BIN=/usr/local/bin/veilkey-cli
ENV VEILKEY_VK_BIN=/usr/local/bin/veilkey
ENV VEILKEY_SESSION_CONFIG_BIN=/usr/local/bin/veilkey-session-config
ENV VEILKEY_CONFIG=/root/.veilkey.yml

WORKDIR /root
ENTRYPOINT ["/bin/bash"]
