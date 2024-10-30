FROM alpine:3 as pluginbuilder
ARG PLUGIN_MODULE=github.com/ret2binsh/traefik-plugin-permituseragent
ARG PLUGIN_GIT_REPO=https://github.com/ret2binsh/traefik-plugin-permituseragent.git
ARG PLUGIN_GIT_BRANCH=main
RUN apk add --update git && \
    git clone ${PLUGIN_GIT_REPO} /plugins-local/src/${PLUGIN_MODULE} \
      --depth 1 --single-branch --branch ${PLUGIN_GIT_BRANCH}

FROM traefik:latest
COPY --from=pluginbuilder /plugins-local /plugins-local
