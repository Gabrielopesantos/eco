FROM scratch as cache

COPY bin .

FROM scratch as ship

ARG TARGETOS
ARG TARGETARCH
ARG BIN_NAME

COPY --from=cache /${TARGETOS}_${TARGETARCH}/${BIN_NAME} ./eco

ENTRYPOINT ["/eco"]
