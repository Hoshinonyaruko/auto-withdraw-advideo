package template

const ConfigTemplate = `
version: 1
settings:

  #通用配置项
  port : "28800"                                #本程序监听的端口
  wspath : ""                                   #ws服务器要监听的端点,默认是直接监听:port
  wstoken : ""                                  #ws服务器的token
  paths : []                                    #当要连接多个onebotv11的http正向地址时,多个地址填入这里.
  video_second_limit : 5                        #低于5秒的视频就会被撤回.
  check_video_qrcode : true                     #检查低于n秒的视频是否存在qr码.有则撤回.
  qr_limit : 1                                  #逐帧检查视频,包含1帧二维码就撤回.
  access_tokens:
  - self_id: ""
    token: ""
`
