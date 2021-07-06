## 使用说明

#### 功能介绍

- 内网环境下的Http服务，无法在外网访问，可以通过此工具完成外网访问内网应用
- 常用场景
  - 微信，支付宝（小程序，支付等）开放文档的对接时，需要本地调试
  - 和其他第三方联调时等等
  - 内网的测试服务，需要给到外网，让其他开发或测试更方便访问

- 支持功能
  - 支持1对1,N对N的穿透功能，如果您拥有域名，则可以无限制部署穿透通道
  - 支持Https协议，如果您进行支付测试，很有可能遇到必须https协议的限制，这个完全可用
  - 配置简单，同一个执行文件即可完成服务端和客户端的部署

#### 文件说明
使用的话，只用关注下面三个文件

`bin/natapp.sh` : linux下使用的执行文件

`bing/natapp.exe`: window下使用的执行文件*(本质和natapp.sh一样,只是不同操作系统使用)*

`client.json`: 客户端的配置文件



#### 部署说明（样例）

>  远程外网机器为Linux系统.    假设IP为: 1.2.3.4

>  本地内网机器为window系统,可用访问的服务两个：**192.168.1.11:8080**/**192.168.1.12:8888**



#### 一、启动服务端

1.将文件上传到服务器,并修改权限

```shell
chmod 755 natapp.sh
```

2.进入natapp.sh所在目录,直接执行命令

```shell
./natapp.sh -m server -h 8080 -t 18888

#参数说明:
-m : 启动模式, server服务端,client客户端,默认为client
-h : 域名访问HTTP端口,默认8080，确保端口开放可外网访问
-t : 客户端链接TCP端口,默认18888 确保端口开放可外网访问

#https参数说明,非必须配置项，可以不用配置
-tls : 是否带有http证书,启动端口为443，证书最好购买CA机构官方的通配符证书，可以多个二级域名同时使用
-pem ：证书路径,默认server.pem
-key ：服务端私钥,默认server.key

#完整例子
./natapp.sh -m server -h 80 -t 18888 -tls -pem cert/server.pem -key cert/server.key
```



#### 二、启动客户端

1. 准备好配置文件 `client.json`和`natapp.exe`同一个目录

   ```json
   #样例配置,配置文件中不要加注释,直接配置json文件
   #根据ip解析的域名,可以本地转发到更多通道,具体样例配置如下:
   {
     "serverHost": "1.2.3.4:18888",
     "clients": [
       {"domain": "a.test.com:8080","forward":"192.168.1.11:8080"},
       {"domain": "b.test.com:8080","forward":"192.168.1.12:8888"}
     ]
   }
   ```

- **serverHost:** 远程服务器的ip的地址,对应服务端端口是**-t**参数指定的端口

  ```JSON
  "serverHost": "1.2.3.4:18888"
  ```

- **clients:** 转发列表

- **domain:** 外网域名地址,该地址为域名解析到外网服务器IP,
  
- **forward:** 本地内网地址	

  假设您有域名**test.com**，配置如下:

  - **a.test.com**   解析IP到        **1.2.3.4**

  - **b.test.com**   也解析IP到   **1.2.3.4**

    ```json
    "clients": [
       {"domain": "a.test.com:8080","forward":"192.168.1.11:8080"},
       {"domain": "b.test.com:8080","forward":"192.168.1.12:8888"}
    ]
    ```

  如果您没有域名

  - 则无法使用多通道，ip只能部署1个

    ```JSON
    "clients": [
      {"domain": "1.2.3.4:8080","forward":"192.168.1.12:8888"}
    ]
    ```



2. 启动命令

   > 启动之后，由于window10的cmd存在快速编辑功能，需关闭，容易阻塞通道的网络
   
   **双击 natapp.exe**



#### 三、测试访问

- 直接访问 **a.test.com:8080** 则访问到 **192.168.1.11:8080** 应用
- 直接访问 **b.test.com:8080** 则访问到 **192.168.1.12:8888** 应用



#### 原理图

![](image/%E5%BE%AE%E4%BF%A1%E5%9B%BE%E7%89%87_20210702174755.png)

#### 配套启动shell脚本（服务端）

- 后台启动

```shell
nohup ./natapp.sh -m server -h 8080 -t 18888 &
tail -f nohup.out
```

- 停止命令

```shell
#!/bin/bash
pid=`ps -ef | grep natapp |grep -v grep | awk '{print $2}'`
if [ ${pid} ]; then
	kill -15 ${pid}
if
```

