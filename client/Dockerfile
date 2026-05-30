# Build stage
FROM node:24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package.json (package-lock.json excluded via .dockerignore to avoid private registry URLs)
COPY package.json ./

# Install ALL dependencies from public npm registry (need devDeps for build)
RUN npm config set registry https://registry.npmjs.org/ && \
    npm install

# Copy source code
COPY . .

# Build the application (Vite outputs to dist/)
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy built files from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Expose port
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
