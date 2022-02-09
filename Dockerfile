ARG ARCH

FROM multiarch/alpine:${ARCH}-latest-stable

ARG USER=shhh
ENV HOME /home/${USER}

RUN apk add --no-cache bash
RUN adduser -D -s /bin/bash ${USER}

USER ${USER}
WORKDIR ${HOME}

# binary
COPY shhh /usr/local/bin/

# shhh variables
ENV SHHH_FILENAME ""
ENV SHHH_PASSWD ""
ENV SHHH_GROUP ""
ENV SHHH_KEY ""
ENV SHHH_VALUE ""

# bash prompt
ENV PS1 "$(whoami):$(pwd)> "

ENTRYPOINT ["/usr/local/bin/shhh"]
CMD [ "--help" ]