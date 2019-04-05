# Forwardhook

转发webhook请求到多后端服务

## 使用

```bash
./forwardhook -c ./config.json
```

## 配置文件
```json
{
    "listen": "0.0.0.0:8899", //本地监听地址
    "retries": 10, //后端最大重试次数（以返回HTTP 200状态为准）
    "mappings": [ //路径映射配置表
        {
            "path": "/hello/", //匹配路径
            "sites": [
                "http://www.google.com/hello" //转发后端路径
            ]
        },
        {
            "path": "/test/",
            "sites": [
                "http://www.google.com/test"
            ]
        }
    ]
}
```


### How to build the docker container

This is based on the [minimal docker container](http://blog.codeship.com/building-minimal-docker-containers-for-go-applications/) article from Codeship.

SSL certificates are bundled in to get around x509 errors when requesting SSL
endpoints.

```bash
docker build -t --rm mm-sam/forwardhook -f Dockerfile.scratch .

# Push to docker hub
docker push bittersweet/forwardhook
```

## Run it locally

```bash
docker run -e "FORWARDHOOK_SITES=https://site:port/path" --rm -p 8000:8000 -it bittersweet/forwardhook
curl local.docker:8000
```
