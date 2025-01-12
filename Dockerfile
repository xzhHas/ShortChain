FROM alpine

COPY . .

# 设置工作目录为 /app
WORKDIR .

EXPOSE 8000
EXPOSE 9000

# 声明容器运行时可挂载的配置卷（用于存放配置文件）
VOLUME /data/conf

# 设置容器启动时的默认命令，启动 main 可执行文件并指定配置路径
CMD ["./cmd/shortUrlX/shortUrlX", "-conf", "/data/conf"]

# docker build -t short .
# docker run -d -p 8000:8000 -p 9000:9000 --name shortchain short