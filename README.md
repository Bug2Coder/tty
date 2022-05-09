# tty
#### web终端的后台支持模块

#### 本模块功能：
##### 1、支持本地虚拟终端的构建
##### 2、为web终端提供后台解析支持、（全功能支持终端功能、会话启动目录为项目目录）
##### 3、一个链接对应一个终端
##### 4、链接和会话均可手动关闭、即web端输入`exit`退出和调用端主动关闭会话
#### 使用说明：
##### 1、创建虚拟终端会话需要传入满足`io.ReadWriteCloser`接口的类型、需要自行封装满足的类型
##### 2、读取接口读取到的数据必须为最终终端数据，不能包含其他数据、如需封装其他业务、请自行封装类型并解析拆分后传入本模块中
