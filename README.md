# rtcdctun
基于rtc的tunnel
server -dial 127.0.0.1:25565

client -listen 127.0.0.1:25566
会生成offer报文，粘贴到server那里，server会回应answer报文
粘贴answer报文，开始连接
