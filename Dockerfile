FROM alpine
RUN apk --no-cache add curl
ADD  product /product
ENTRYPOINT [ "/product" ]