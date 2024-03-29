FROM  golang:1.17.7 as builder

MAINTAINER ybyang7@iflytek.com
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
RUN chmod +x /root/${BIN_NAME}
ENTRYPOINT ["/root/avp_linux_amd64"]
CMD ["server"]
