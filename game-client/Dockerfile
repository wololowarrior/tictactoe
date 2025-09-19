# Use nginx to serve static files
FROM nginx:alpine

# Copy only static files (not server.js since we're using nginx)
COPY index.html /usr/share/nginx/html/
COPY game.js /usr/share/nginx/html/
COPY nakama-browser.js /usr/share/nginx/html/

# Replace default nginx server config
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
