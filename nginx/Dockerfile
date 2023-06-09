FROM nginx:latest
EXPOSE 80
RUN mkdir -p "app/static/css" && mkdir "app/static/img" && mkdir "app/static/js"
COPY nginx.conf /etc/nginx/nginx.conf