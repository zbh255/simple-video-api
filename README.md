# Simple-video-api
简单实现的基于ffmpeg的视频剪辑在线接口
## Description
无需`Mysql`和`Redis`等额外组件实现的业务代码，拷贝了自己实现的goroutine池，
LittleHeap + HashMap实现了一个高效的根据过期时间管理User Token的Token Pool
## Video API
### WebSocket Call FFmpeg
WebSocket升级之前需要先通过Token校验用户身份，需要通过`HTTP Header`携带
> Authentication Bearer $User-Token

> 所有查询类型的接口的响应均为以下格式
```json
{
  "Code": 200,
  "Msg": "Ok"
}
```
请求调用视频剪辑接口需要的参数
```json
{
  "type": "call",
  "video_url": "http://host:port/video",
  "video_out_type": "mp4",
  "video_start": "00:01:04",
  "video_end": "00:01:59"
}
```
`type` == call表示为这是一次调用的请求,`check_status`表示查询处理进度
请求被提交的回复
- `type`表示消息的类型，回复的type均为`return`
- `task_id`表示本次处理的任务id，这只是一个暂时的存储，之后可以通过此来查询处理进度
- `state_comment`只是一段对本次的回复的一些注释
- `split_time`表示根据end-start的长度
- `split_video_id`是一个具体的文件名，这是一个持久的存储，处理好的视频将根据此来命名，但是不可以用来查询
```json
{
  "type": "return",
  "task_id": 1024,
  "state_comment": "视频处理任务已被提交",
  "split_time": "40s",
  "split_video_id": "f069112e-ef33-11ec-af07-54b203942810.mp4"
}
```
我将业务逻辑设定为在WebSocket连接上提交剪辑任务并不会阻塞当前连接，所以你可以在返回后可以选择发送查询
处理状态的消息或者去做其它事情
要查询一个任务的进度需要提供的消息如下
```json
{
  "type": "check_status",
  "task_id": 1024
}
```
一个未处理完的回复如下
```json
{
  "Code" : 10022,
  "Msg" : "视频任务正在处理"
}
```
### WebSocket连接外的查询接口
根据`task_id`来查询进度
> GET /video/task_id

Example
> GET /video/1024

获取已经剪辑完毕的视频
> GET /video/source/split_video_id

Example
> GET /video/source/f069112e-ef33-11ec-af07-54b203942810.mp4

## User API
### Sign
HTTP Header
> POST /user/sign

HTTP Body
```json
{
  "user_name": "hh",
  "user_password": ""
}
```
Ok Response
```json
{
  "Code": 200,
  "Msg": "ok"
}
```
Error Response,The Code is no 200
```json
{
  "Code": 10010,
  "Msg": "json参数不正确"
}
```
### Login
HTTP Header
> POST /user/login

HTTP Body
```json
{
  "user_name": "hh",
  "user_password": ""
}
```
Ok Response
```json
{
  "Code": 200,
  "Msg": "ok",
  "token": "f069112e-ef33-11ec-af07-54b203942810"
}
```
Error Response,The Code is no 200
```json
{
  "Code": 10010,
  "Msg": "json参数不正确"
}
```