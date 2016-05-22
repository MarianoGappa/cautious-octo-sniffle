FROM alpine:3.3

ADD flowbro /flowbro
ADD webroot /webroot

EXPOSE 41234

ENTRYPOINT ["/flowbro"]
