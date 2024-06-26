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
  set_group_kick : false                        #检测到符合条件的视频在撤回后踢掉发送者.
  kick_and_reject_add_request : false           #踢掉后禁止再次加群
  qr_limit : 1                                  #逐帧检查视频,包含1帧二维码就撤回.
  withdraw_notice : "撤回了一条广告."                          #撤回广告时的回复.
  on_enable_video_check : "视频广告撤回on"       #视频二维码广告撤回开启指令(默认关闭)需手动发指令开启
  on_disable_video_check : "视频广告撤回off"     #视频二维码广告撤回关闭指令
  on_enable_pic_check : "图片广告撤回on"         #图片二维码广告撤回开启指令(默认关闭)需手动发指令开启
  on_disable_pic_check : "图片广告撤回off"       #图片二维码广告撤回关闭指令
  withdraw_words : [""]                         #该配置无开关,请将你最讨厌的广告关键词放进去,比如"免费收徒\抖音引流\保证一天",检测到就会自动撤回
  access_tokens:
  - self_id: ""
    token: ""
`
