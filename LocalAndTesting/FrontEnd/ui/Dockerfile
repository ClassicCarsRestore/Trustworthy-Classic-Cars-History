# Stage 1: Build the Angular app
FROM node:20.3.1 AS build
RUN mkdir -p /app
WORKDIR /app
COPY package.json /app
RUN npm install
COPY . /app
RUN npm run build --prod

# Stage 2: Serve the Angular app using NGINX
FROM nginx:alpine
COPY --from=build /app/dist/ui /usr/share/nginx/html
COPY nginxConf/default.conf /etc/nginx/conf.d/default.conf