# --- 构建前端 ---
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
# 复制依赖文件
COPY frontend/package*.json ./
RUN npm install
# 复制源码构建
COPY frontend/ .
RUN npm run build

# --- 构建后端 ---
FROM golang:1.25.1-alpine AS backend-builder
WORKDIR /app
COPY go.mod go.sum .
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod tidy
COPY . .
RUN go build -o main main.go

# --- 运行环境 ---
FROM nginx:1.25.3-alpine
# 安装运行 Go 程序所需的库（alpine 基础镜像可能需要）
RUN apk add --no-cache libc6-compat 

COPY --from=backend-builder /app/main /app/main
COPY --from=backend-builder /app/config.yaml /app/config.yaml
COPY --from=frontend-builder /app/frontend/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf

WORKDIR /app
EXPOSE 80
CMD ["/bin/sh", "-c", "./main & nginx -g 'daemon off;'"]
