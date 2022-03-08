FROM  golang:1.17.7 as builder


WORKDIR /workspace

COPY . .

RUN make

FROM alpine:latest as prod
ENV TZ Asia/Shanghai
ENV BIN_NAME avp_linux_amd64
RUN set -eux && sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add tzdata && cp /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone
WORKDIR /root/
COPY --from=0 /workspace/bin/linux .
#COPY --from=docbuilder /workspace/website/public ./website/public
CMD ["/root/${BIN_NAME}"]